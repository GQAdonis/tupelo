//go:generate msgp

package remote

import (
	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
	"github.com/tinylib/msgp/msgp"
)

type remoteDeliver struct {
	header       actor.ReadonlyMessageHeader
	message      interface{}
	target       *actor.PID
	sender       *actor.PID
	serializerID int32
}

func ToWireDelivery(rd *remoteDeliver) *WireDelivery {
	marshaled, err := rd.message.(msgp.Marshaler).MarshalMsg(nil)
	if err != nil {
		panic(fmt.Errorf("could not marshal message: %v", err))
	}
	wd := &WireDelivery{
		Message: marshaled,
		Type:    messages.GetTypeCode(rd.message),
		Target:  ToActorPid(rd.target),
		Sender:  ToActorPid(rd.sender),
	}
	if rd.header != nil {
		wd.Header = rd.header.ToMap()
	}
	return wd
}

type registerBridge struct {
	Peer    string
	Handler *actor.PID
}

type ActorPID struct {
	Address string
	Id      string
}

func ToActorPid(a *actor.PID) *ActorPID {
	if a == nil {
		return nil
	}
	return &ActorPID{
		Address: a.Address,
		Id:      a.Id,
	}
}

func FromActorPid(a *ActorPID) *actor.PID {
	return actor.NewPID(a.Address, a.Id)
}

type WireDelivery struct {
	Header   map[string]string
	Message  []byte
	Type     int8
	Target   *ActorPID
	Sender   *ActorPID
	Outgoing bool `msg:"-"`
}

func (wd *WireDelivery) GetMessage() (msgp.Unmarshaler, error) {
	msg := messages.GetUnmarshaler(wd.Type)
	_, err := msg.UnmarshalMsg(wd.Message)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling message: %v", err)
	}
	return msg, nil
}