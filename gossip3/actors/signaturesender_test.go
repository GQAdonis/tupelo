package actors

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/quorumcontrol/messages/v2/build/go/signatures"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/types"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
	"github.com/quorumcontrol/tupelo/testnotarygroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendSigs(t *testing.T) {
	ts := testnotarygroup.NewTestSet(t, 1)
	rootContext := actor.EmptyRootContext
	ss := rootContext.Spawn(NewSignatureSenderProps())
	defer rootContext.Poison(ss)

	fut := actor.NewFuture(5 * time.Second)
	subscriberFunc := func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *signatures.TreeState:
			context.Send(fut.PID(), msg)
		}
	}

	subscriber := rootContext.Spawn(actor.PropsFromFunc(subscriberFunc))
	defer rootContext.Poison(subscriber)

	signer := types.NewLocalSigner(ts.PubKeys[0], ts.SignKeys[0])
	signer.Actor = subscriber

	rootContext.Send(ss, &messages.SignatureWrapper{
		State:            &signatures.TreeState{TransactionId: []byte("testonly")},
		RewardsCommittee: []*types.Signer{signer},
	})

	msg, err := fut.Result()
	require.Nil(t, err)
	assert.Equal(t, []byte("testonly"), msg.(*signatures.TreeState).TransactionId)
}
