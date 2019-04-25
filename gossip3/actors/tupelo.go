package actors

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/plugin"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/quorumcontrol/storage"
	"github.com/quorumcontrol/tupelo-go-client/client"
	extmsgs "github.com/quorumcontrol/tupelo-go-client/gossip3/messages"
	"github.com/quorumcontrol/tupelo-go-client/gossip3/middleware"
	"github.com/quorumcontrol/tupelo-go-client/gossip3/remote"
	"github.com/quorumcontrol/tupelo-go-client/gossip3/types"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
)

const commitPubSubTopic = "tupelo-commits"

const ErrBadTransaction = 1

// TupeloNode is the main logic of the entire system,
// consisting of multiple gossipers
type TupeloNode struct {
	middleware.LogAwareHolder

	self              *types.Signer
	notaryGroup       *types.NotaryGroup
	conflictSetRouter *actor.PID
	validatorPool     *actor.PID
	cfg               *TupeloConfig
}

type TupeloConfig struct {
	Self              *types.Signer
	NotaryGroup       *types.NotaryGroup
	CurrentStateStore storage.Storage
	PubSubSystem      remote.PubSub
}

func NewTupeloNodeProps(cfg *TupeloConfig) *actor.Props {
	return actor.PropsFromProducer(func() actor.Actor {
		return &TupeloNode{
			self:        cfg.Self,
			notaryGroup: cfg.NotaryGroup,
			cfg:         cfg,
		}
	}).WithReceiverMiddleware(
		middleware.LoggingMiddleware,
		plugin.Use(&middleware.LogPlugin{}),
	)
}

func (tn *TupeloNode) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		tn.handleStarted(context)
	case *extmsgs.GetTip:
		tn.handleGetTip(context, msg)
	case *messages.CurrentStateWrapper:
		tn.handleNewCurrentStateWrapper(context, msg)
	case *extmsgs.CurrentState:
		context.Forward(tn.conflictSetRouter)
	case *extmsgs.Signature:
		context.Forward(tn.conflictSetRouter)
	case *extmsgs.Transaction:
		tn.handleNewTransaction(context)
	case *messages.ValidateTransaction:
		tn.handleNewTransaction(context)
	case *messages.TransactionWrapper:
		tn.handleNewTransaction(context)
	}
}

func (tn *TupeloNode) handleNewCurrentStateWrapper(context actor.Context, msg *messages.CurrentStateWrapper) {
	if msg.Verified {
		tn.Log.Infow("commit", "tx", msg.CurrentState.Signature.TransactionID, "seen", msg.Metadata["seen"])
		err := tn.cfg.CurrentStateStore.Set(msg.CurrentState.CurrentKey(), msg.MustMarshal())
		if err != nil {
			panic(fmt.Errorf("error setting current state: %v", err))
		}
		tn.Log.Debugw("tupelo node sending activatesnoozingconflictsets", "objectID", msg.CurrentState.Signature.ObjectID)
		// un-snooze waiting conflict sets
		context.Send(tn.conflictSetRouter, &messages.ActivateSnoozingConflictSets{ObjectID: msg.CurrentState.Signature.ObjectID})

		// if we are the ones creating this current state then broadcast
		if msg.Internal {
			tn.Log.Debugw("publishing new current state", "topic", string(msg.CurrentState.Signature.ObjectID))
			if err := tn.cfg.PubSubSystem.Broadcast(string(msg.CurrentState.Signature.ObjectID), msg.CurrentState); err != nil {
				tn.Log.Errorw("error publishing", "err", err)
			}
			if err := tn.cfg.PubSubSystem.Broadcast(commitPubSubTopic, msg.CurrentState); err != nil {
				tn.Log.Errorw("error publishing", "err", err)
			}

			for _, transWrapper := range msg.FailedTransactions {
				tn.Log.Debugw("publishing failed transaction", "tx", transWrapper.TransactionID)
				err := tn.cfg.PubSubSystem.Broadcast(string(transWrapper.Transaction.ObjectID), &extmsgs.Error{
					Source: string(transWrapper.TransactionID),
					Code:   ErrBadTransaction,
					Memo:   fmt.Sprintf("bad transaction"),
				})
				if err != nil {
					tn.Log.Errorw("error publishing", "err", err)
				}
			}
		}
	}
}

func (tn *TupeloNode) handleNewTransaction(context actor.Context) {
	switch msg := context.Message().(type) {
	case *extmsgs.Transaction:
		// broadcaster has sent us a fresh transaction
		tn.validateTransaction(context, &messages.ValidateTransaction{
			Transaction: msg,
		})
	case *messages.ValidateTransaction:
		// snoozed transaction has been activated and needs full validation
		tn.validateTransaction(context, msg)
	case *messages.TransactionWrapper:
		// validatorPool has validated or rejected a transaction we sent it above
		if msg.PreFlight || msg.Accepted {
			context.Send(tn.conflictSetRouter, msg)
		} else {
			if msg.Stale {
				tn.Log.Debugw("ignoring and cleaning up stale transaction", "msg", msg)
				err := tn.cfg.PubSubSystem.Broadcast(string(msg.Transaction.ObjectID), &extmsgs.Error{
					Source: string(msg.TransactionID),
					Code:   ErrBadTransaction,
					Memo:   "stale",
				})
				if err != nil {
					tn.Log.Errorw("error publishing", "err", err)
				}
			} else {
				tn.Log.Debugw("removing bad transaction", "msg", msg)
				err := tn.cfg.PubSubSystem.Broadcast(string(msg.Transaction.ObjectID), &extmsgs.Error{
					Source: string(msg.TransactionID),
					Code:   ErrBadTransaction,
					Memo:   fmt.Sprintf("bad transaction: %v", msg.Metadata["error"]),
				})
				if err != nil {
					tn.Log.Errorw("error publishing", "err", err)
				}
			}
			msg.StopTrace()
		}
	}
}

func (tn *TupeloNode) validateTransaction(context actor.Context, msg *messages.ValidateTransaction) {
	tn.Log.Debugw("validating transaction", "msg", msg)
	context.Request(tn.validatorPool, &validationRequest{
		transaction: msg.Transaction,
	})
}

func (tn *TupeloNode) handleGetTip(context actor.Context, msg *extmsgs.GetTip) {
	tn.Log.Debugw("handleGetTip", "tip", msg.ObjectID)
	currStateBits, err := tn.cfg.CurrentStateStore.Get(msg.ObjectID)
	if err != nil {
		panic(fmt.Errorf("error getting tip: %v", err))
	}

	var currState extmsgs.CurrentState

	if len(currStateBits) > 0 {
		_, err = currState.UnmarshalMsg(currStateBits)
		if err != nil {
			panic(fmt.Errorf("error unmarshaling CurrentState: %v", err))
		}
	}

	context.Respond(&currState)
}

func (tn *TupeloNode) handleStarted(context actor.Context) {
	sender, err := context.SpawnNamed(NewSignatureSenderProps(), "signatureSender")
	if err != nil {
		panic(fmt.Sprintf("error spawning: %v", err))
	}

	sigGenerator, err := context.SpawnNamed(NewSignatureGeneratorProps(tn.self, tn.notaryGroup), "signatureGenerator")
	if err != nil {
		panic(fmt.Sprintf("error spawning: %v", err))
	}

	sigChecker, err := context.SpawnNamed(NewSignatureVerifier(), "sigChecker")
	if err != nil {
		panic(fmt.Sprintf("error spawning: %v", err))
	}

	_, err = context.SpawnNamed(tn.cfg.PubSubSystem.NewSubscriberProps(client.TransactionBroadcastTopic), "broadcast-subscriber")
	if err != nil {
		panic(fmt.Sprintf("error spawning broadcast receiver: %v", err))
	}

	topicValidator := newCommitValidator(tn.cfg.NotaryGroup, sigChecker)
	err = tn.cfg.PubSubSystem.RegisterTopicValidator(commitPubSubTopic, topicValidator.validate, pubsub.WithValidatorTimeout(500*time.Millisecond), pubsub.WithValidatorConcurrency(verifierConcurrency*2))
	if err != nil {
		panic(fmt.Sprintf("error registering topic validator: %v", err))
	}

	_, err = context.SpawnNamed(tn.cfg.PubSubSystem.NewSubscriberProps(commitPubSubTopic), "commit-subscriber")
	if err != nil {
		panic(fmt.Sprintf("error spawning commit receiver: %v", err))
	}

	tvConfig := &TransactionValidatorConfig{
		NotaryGroup:       tn.notaryGroup,
		SignatureChecker:  sigChecker,
		CurrentStateStore: tn.cfg.CurrentStateStore,
	}
	validatorPool, err := context.SpawnNamed(NewTransactionValidatorProps(tvConfig), "validator")
	if err != nil {
		panic(fmt.Sprintf("error spawning: %v", err))
	}

	csrConfig := &ConflictSetRouterConfig{
		NotaryGroup:        tn.notaryGroup,
		Signer:             tn.self,
		SignatureGenerator: sigGenerator,
		SignatureChecker:   sigChecker,
		SignatureSender:    sender,
		CurrentStateStore:  tn.cfg.CurrentStateStore,
	}
	router, err := context.SpawnNamed(NewConflictSetRouterProps(csrConfig), "conflictSetRouter")
	if err != nil {
		panic(fmt.Sprintf("error spawning: %v", err))
	}

	tn.conflictSetRouter = router
	tn.validatorPool = validatorPool
}
