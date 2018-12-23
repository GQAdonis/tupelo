package actors

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/Workiva/go-datastructures/bitarray"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
	"github.com/quorumcontrol/tupelo/gossip3/types"
	"github.com/quorumcontrol/tupelo/testnotarygroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureGenerator(t *testing.T) {
	ts := testnotarygroup.NewTestSet(t, 1)
	signer := types.NewLocalSigner(ts.PubKeys[0].ToEcdsaPub(), ts.SignKeys[0])
	ng := types.NewNotaryGroup()
	ng.AddSigner(signer)
	currentState := actor.Spawn(NewStorageProps())
	defer currentState.Poison()
	validator := actor.Spawn(NewTransactionValidatorProps(currentState))
	defer validator.Poison()

	sigGnerator := actor.Spawn(NewSignatureGeneratorProps(signer, ng))
	defer sigGnerator.Poison()

	fut := actor.NewFuture(5 * time.Second)
	validatorSenderFunc := func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *messages.Store:
			context.Request(validator, msg)
		case *messages.SignatureWrapper:
			fut.PID().Tell(msg)
		case *messages.TransactionWrapper:
			context.Request(sigGnerator, msg)
		}
	}

	sender := actor.Spawn(actor.FromFunc(validatorSenderFunc))
	defer sender.Poison()

	trans := newValidTransaction(t)
	value, err := trans.MarshalMsg(nil)
	require.Nil(t, err)
	key := crypto.Keccak256(value)

	sender.Tell(&messages.Store{
		Key:   key,
		Value: value,
	})

	msg, err := fut.Result()
	require.Nil(t, err)

	sigWrapper := msg.(*messages.SignatureWrapper)
	arry, err := bitarray.Unmarshal(sigWrapper.Signature.Signers)
	require.Nil(t, err)
	isSet, err := arry.GetBit(0)
	require.Nil(t, err)
	assert.True(t, isSet)
}
