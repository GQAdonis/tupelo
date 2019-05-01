package actors

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/plugin"
	"github.com/AsynkronIT/protoactor-go/router"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/middleware"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
)

type SignatureSender struct {
	middleware.LogAwareHolder
}

const senderConcurrency = 100

func NewSignatureSenderProps() *actor.Props {
	return router.NewRoundRobinPool(senderConcurrency).WithProducer(func() actor.Actor {
		return new(SignatureSender)
	}).WithReceiverMiddleware(
		middleware.LoggingMiddleware,
		plugin.Use(&middleware.LogPlugin{}),
	)
}

func (ss *SignatureSender) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.SignatureWrapper:
		for _, target := range msg.RewardsCommittee {
			ss.Log.Debugw("sending", "t", target.ID, "actor", target.Actor)
			context.Send(target.Actor, msg.Signature)
		}
	}
}
