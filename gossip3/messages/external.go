//go:generate msgp

package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/quorumcontrol/differencedigest/ibf"
)

func init() {
	RegisterEncodable(Ping{})
	RegisterEncodable(Pong{})
	RegisterEncodable(Store{})
	RegisterEncodable(GetSyncer{})
	RegisterEncodable(SyncDone{})
	RegisterEncodable(NoSyncersAvailable{})
	RegisterEncodable(SyncerAvailable{})
	RegisterEncodable(CurrentState{})
	RegisterEncodable(Signature{})
	RegisterEncodable(Transaction{})
	RegisterEncodable(GetTip{})
	RegisterEncodable(ActorPID{})
	RegisterEncodable(ProvideStrata{})
	RegisterEncodable(ProvideBloomFilter{})
	RegisterEncodable(RequestKeys{})
}

type DestinationHolder struct {
	Destination *ActorPID
}

func (dh *DestinationHolder) SetDestination(newDestination *ActorPID) {
	dh.Destination = newDestination
}

func (dh *DestinationHolder) GetDestination() *ActorPID {
	return dh.Destination
}

type DestinationSettable interface {
	SetDestination(*ActorPID)
	GetDestination() *ActorPID
}

type Ping struct {
	Msg string
}

type Pong struct {
	Msg string
}

type Store struct {
	Key   []byte
	Value []byte
}

type GetSyncer struct {
	Kind string
}

type SyncDone struct{}

type NoSyncersAvailable struct{}

type SyncerAvailable struct {
	DestinationHolder
}

type CurrentState struct {
	ObjectID  []byte
	Tip       []byte
	OldTip    []byte // This is a big deal, worth talking through
	Signature Signature
}

func (cs *CurrentState) StorageKey() []byte {
	if len(cs.OldTip) == 0 {
		return append(cs.ObjectID, byte(0))
	}
	return append(cs.ObjectID, cs.OldTip...)

}

func (cs *CurrentState) CurrentKey() []byte {
	return append(cs.ObjectID)
}

func (cs *CurrentState) MustBytes() []byte {
	bits, err := cs.MarshalMsg(nil)
	if err != nil {
		panic(fmt.Errorf("error marshaling current state: %v", err))
	}
	return bits
}

type GetTip struct {
	ObjectID []byte
}

type RequestKeys struct {
	Keys []uint64
}

type Signature struct {
	TransactionID []byte
	ObjectID      []byte
	PreviousTip   []byte
	NewTip        []byte
	View          uint64
	Cycle         uint64
	Signers       []byte // this is a marshaled BitArray from github.com/Workiva/go-datastructures
	Signature     []byte
}

// type SignatureWrapper struct {
// 	Internal      bool
// 	ConflictSetID string
// 	Signers       SignerMap
// 	Signature     *Signature
// 	Metadata      MetadataMap
// }

func (sig *Signature) GetSignable() []byte {
	return append(append(sig.ObjectID, append(sig.PreviousTip, sig.NewTip...)...), append(uint64ToBytes(sig.View), uint64ToBytes(sig.Cycle)...)...)
}

func uint64ToBytes(id uint64) []byte {
	a := make([]byte, 8)
	binary.BigEndian.PutUint64(a, id)
	return a
}

type Transaction struct {
	ObjectID    []byte
	PreviousTip []byte
	NewTip      []byte
	Payload     []byte
}

func (t *Transaction) ConflictSetID() string {
	return ConflictSetID(t.ObjectID, t.PreviousTip)
}

func ConflictSetID(objectID, previousTip []byte) string {
	return string(append(objectID, previousTip...))
}

type ProvideStrata struct {
	DestinationHolder

	Strata *ibf.DifferenceStrata
}

type ProvideBloomFilter struct {
	DestinationHolder

	Filter *ibf.InvertibleBloomFilter
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
