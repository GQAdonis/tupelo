package actors

import (
	"github.com/quorumcontrol/tupelo-go-sdk/consensus"
	"github.com/quorumcontrol/messages/build/go/signatures"
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/quorumcontrol/tupelo-go-sdk/bls"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/plugin"
	"github.com/AsynkronIT/protoactor-go/router"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/middleware"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/types"
	"github.com/quorumcontrol/tupelo-go-sdk/tracing"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
)

type signatureByTransaction map[string]*messages.SignatureWrapper
type transactionMap map[string]*messages.TransactionWrapper

type checkStateMsg struct {
	atUpdate uint64
}

type csWorkerRequest struct {
	msg interface{}
	cs  *ConflictSet
}

// implements the necessary interface for consistent hashing
// router
func (cswr *csWorkerRequest) Hash() string {
	return cswr.cs.ID
}

type ConflictSet struct {
	tracing.ContextHolder

	ID string

	done          bool
	signatures    signatureByTransaction
	didSign       bool
	transactions  transactionMap
	snoozedCommit *messages.CurrentStateWrapper
	// a map of signers that have been part of a seen signature for this
	// view
	hasSignedSomething map[string]struct{}
	view               uint64
	updates            uint64
	active             bool
}

type ConflictSetConfig struct {
	NotaryGroup        *types.NotaryGroup
	Signer             *types.Signer
	SignatureGenerator *actor.PID
	SignatureSender    *actor.PID
	ConflictSetRouter  *actor.PID
	CommitValidator    *commitValidator
}

func NewConflictSet(id string) *ConflictSet {
	return &ConflictSet{
		ID:                 id,
		signatures:         make(signatureByTransaction),
		transactions:       make(transactionMap),
		hasSignedSomething: make(map[string]struct{}),
	}
}

const conflictSetConcurrency = 50

func NewConflictSetWorkerPool(cfg *ConflictSetConfig) *actor.Props {
	// it's important that this is a consistent hash pool rather than round robin
	// because we do not want two operations on a single conflictset executing concurrently
	// if you change this, make sure you provide some other "locking" mechanism.
	return router.NewConsistentHashPool(conflictSetConcurrency).WithProducer(func() actor.Actor {
		return &ConflictSetWorker{
			router:             cfg.ConflictSetRouter,
			notaryGroup:        cfg.NotaryGroup,
			signer:             cfg.Signer,
			signatureGenerator: cfg.SignatureGenerator,
			signatureSender:    cfg.SignatureSender,
			commitValidator:    cfg.CommitValidator,
		}
	}).WithReceiverMiddleware(
		middleware.LoggingMiddleware,
		plugin.Use(&middleware.LogPlugin{}),
	)
}

type ConflictSetWorker struct {
	middleware.LogAwareHolder
	tracing.ContextHolder

	router             *actor.PID
	notaryGroup        *types.NotaryGroup
	signatureGenerator *actor.PID
	signatureSender    *actor.PID
	signer             *types.Signer
	commitValidator    *commitValidator
}

func NewConflictSetWorkerProps(csr *actor.PID) *actor.Props {
	return actor.PropsFromProducer(func() actor.Actor {
		return &ConflictSetWorker{router: csr}
	}).WithReceiverMiddleware(
		middleware.LoggingMiddleware,
		plugin.Use(&middleware.LogPlugin{}),
	)
}

func (csw *ConflictSetWorker) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started, *actor.Stopping, *actor.Stopped:
		// do nothing
	case *csWorkerRequest:
		if msg.cs.done {
			csw.Log.Debugw("received message on done CS")
			return
		}
		csw.dispatchWithConflictSet(msg.cs, msg.msg, context)
	default:
		csw.Log.Errorw("received bad message", "type", reflect.TypeOf(context.Message()).String())
	}
}

func (csw *ConflictSetWorker) dispatchWithConflictSet(cs *ConflictSet, sentMsg interface{}, context actor.Context) {
	switch msg := sentMsg.(type) {
	case *messages.TransactionWrapper:
		csw.handleNewTransaction(cs, context, msg)
	// this will be an external signature
	case *signatures.Signature:
		wrapper, err := sigToWrapper(msg, csw.notaryGroup, csw.signer, false)
		if err != nil {
			panic(fmt.Sprintf("error wrapping sig: %v", err))
		}
		csw.handleNewSignature(cs, context, wrapper)
	case *messages.SignatureWrapper:
		csw.handleNewSignature(cs, context, msg)
	case *messages.CurrentStateWrapper:
		if err := csw.handleCurrentStateWrapper(cs, context, msg); err != nil {
			csw.Log.Errorw("error handling current state", "err", err)
		}
	case *checkStateMsg:
		csw.checkState(cs, context, msg)
	case *messages.ActivateSnoozingConflictSets:
		csw.activate(cs, context, msg)
	default:
		csw.Log.Warnw("received unhandled message type for conflict set", "type",
			reflect.TypeOf(sentMsg).Name())
	}
}

func (csw *ConflictSetWorker) activate(cs *ConflictSet, context actor.Context, msg *messages.ActivateSnoozingConflictSets) {
	sp := cs.NewSpan("activate")
	defer sp.Finish()
	csw.Log.Debug("activate")

	cs.active = true

	if cs.snoozedCommit != nil {
		if err := csw.handleCurrentStateWrapper(cs, context, cs.snoozedCommit); err != nil {
			panic(fmt.Errorf("error processing snoozed commit: %v", err))
		}
	}

	if cs.done {
		// We had a valid commit already, so we're done
		return
	}

	// no (valid) commit, so let's start validating any snoozed transactions
	for _, transaction := range cs.transactions {
		context.Send(csw.router, &messages.ValidateTransaction{
			Transaction: transaction.Transaction,
		})
	}
}

func (csw *ConflictSetWorker) handleNewTransaction(cs *ConflictSet, context actor.Context, msg *messages.TransactionWrapper) {
	sp := cs.NewSpan("handleNewTransaction")
	defer sp.Finish()
	sp.SetTag("transaction", msg.TransactionId)
	sp.SetTag("height", msg.Transaction.Height)
	sp.SetTag("numPrevTransactions", len(cs.transactions))

	transSpan := msg.NewSpan("conflictset-handlenewtransaction")
	defer transSpan.Finish()

	csw.Log.Debugw("new transaction", "trans", msg.TransactionId)
	if !msg.PreFlight && !msg.Accepted {
		panic(fmt.Sprintf("we should only handle pre-flight or accepted transactions at this level"))
	}

	if msg.Accepted {
		sp.SetTag("accepted", true)
		sp.SetTag("active", true)
		transSpan.SetTag("accepted", true)
		cs.active = true
	}

	cs.transactions[string(msg.TransactionId)] = msg
	if cs.active {
		csw.processTransactions(cs, context)
	} else {
		sp.SetTag("snoozing", true)
		transSpan.SetTag("snoozing", true)
		csw.Log.Debugw("snoozing transaction", "t", msg.TransactionId, "height", msg.Transaction.Height)
	}
}

func (csw *ConflictSetWorker) processTransactions(cs *ConflictSet, context actor.Context) {
	sp := cs.NewSpan("processTransactions")
	defer sp.Finish()

	if !cs.active {
		panic(fmt.Errorf("error: processTransactions called on inactive ConflictSet"))
	}

	for _, tw := range cs.transactions {
		transSpan := tw.NewSpan("conflictset-processing")
		csw.Log.Debugw("processing transaction", "t", tw.TransactionId, "height", tw.Transaction.Height)

		if !cs.didSign {
			context.RequestWithCustomSender(csw.signatureGenerator, tw, csw.router)
			transSpan.SetTag("didSign", true)
			cs.didSign = true
		}
		cs.updates++
		transSpan.Finish()
	}
	// do this as a message to make sure we're doing it after all the updates have come in
	context.Send(context.Self(), &csWorkerRequest{cs: cs, msg: &checkStateMsg{atUpdate: cs.updates}})
}

func (csw *ConflictSetWorker) handleNewSignature(cs *ConflictSet, context actor.Context, msg *messages.SignatureWrapper) {
	sp := cs.NewSpan("handleNewSignature")
	defer sp.Finish()

	csw.Log.Debugw("handle new signature", "t", msg.Signature.TransactionId)
	if msg.Internal {
		context.Send(csw.signatureSender, msg)
	}
	if len(msg.Signers) > 1 {
		panic(fmt.Sprintf("currently we don't handle multi signer signatures here"))
	}

	for signer := range msg.Signers {
		cs.hasSignedSomething[signer] = struct{}{}
	}
	//TODO: verify this new sig

	existingSig, ok := cs.signatures[string(msg.Signature.TransactionId)]
	if ok {
		// if the new sig doesn't have any new signatures, then just drop it
		hasNewSigs := false

		for id := range msg.Signers {
			_, ok := existingSig.Signers[id]
			if !ok {
				hasNewSigs = true
				break
			}
		}
		if hasNewSigs {
			csw.Log.Debugw("combining signatures")
			newSig, err := csw.combineSignatures(existingSig, msg)
			if err != nil {
				csw.Log.Infow("error combigning sigs", "err", err)
			}
			csw.Log.Debugw("newsig", "signerCount", len(newSig.Signers), "newSig", newSig)
			cs.signatures[string(msg.Signature.TransactionId)] = newSig
			//TODO: broadcast this new sig
		}
		// else no one cares about this sig, drop it

	} else {
		cs.signatures[string(msg.Signature.TransactionId)] = msg
	}

	cs.updates++
	csw.Log.Debugw("sending checkstate", "self", context.Self().String())
	context.Send(context.Self(), &csWorkerRequest{cs: cs, msg: &checkStateMsg{atUpdate: cs.updates}})
}

func (csw *ConflictSetWorker) checkState(cs *ConflictSet, context actor.Context, msg *checkStateMsg) {
	sp := cs.NewSpan("checkState")
	defer sp.Finish()

	csw.Log.Debugw("check state")
	if cs.updates < msg.atUpdate {
		csw.Log.Debugw("old update")
		sp.SetTag("oldUpdate", true)
		// we know there will be another check state message with a higher update
		return
	}
	if trans := csw.possiblyDone(cs); trans != nil {
		transSpan := trans.NewSpan("checkState")
		defer transSpan.Finish()
		// we have a possibly done transaction, lets make a current state
		if err := csw.createCurrentStateFromTrans(cs, context, trans); err != nil {
			panic(err)
		}
		return
	}

	if csw.deadlocked(cs) {
		csw.handleDeadlockedState(cs, context)
	}
}

func (csw *ConflictSetWorker) handleDeadlockedState(cs *ConflictSet, context actor.Context) {
	sp := cs.NewSpan("handleDeadlockedState")
	defer sp.Finish()
	csw.Log.Debugw("handle deadlocked state")

	var lowestTrans *messages.TransactionWrapper
	for transID, trans := range cs.transactions {
		transSpan := trans.NewSpan("handleDeadlockedState")
		defer transSpan.Finish()
		if lowestTrans == nil {
			lowestTrans = trans
			continue
		}
		if transID < string(lowestTrans.TransactionId) {
			lowestTrans = trans
		}
	}
	cs.nextView(lowestTrans)

	csw.handleNewTransaction(cs, context, lowestTrans)
}

func (cs *ConflictSet) nextView(newWinner *messages.TransactionWrapper) {
	sp := cs.NewSpan("nextView")
	defer sp.Finish()

	cs.view++
	cs.didSign = false
	cs.transactions = make(transactionMap)

	// only keep signatures on the winning transaction
	transSig := cs.signatures[string(newWinner.TransactionId)]
	cs.signatures = signatureByTransaction{string(newWinner.TransactionId): transSig}
	cs.hasSignedSomething = make(map[string]struct{})
	for signerID := range transSig.Signers {
		cs.hasSignedSomething[signerID] = struct{}{}
	}
}

func (csw *ConflictSetWorker) createCurrentStateFromTrans(cs *ConflictSet, actorContext actor.Context, trans *messages.TransactionWrapper) error {
	sp := cs.NewSpan("createCurrentStateFromTrans")
	defer sp.Finish()
	transSpan := trans.NewSpan("createCurrentState")
	defer transSpan.Finish()

	sp.SetTag("winner", trans.TransactionId)

	currState := &signatures.CurrentState{
		Signature: cs.signatures[string(trans.TransactionId)].Signature,
	}

	currStateWrapper := &messages.CurrentStateWrapper{
		Internal:     true,
		Verified:     true, // previously: csw.commitValidator.validate(context.Background(), "", currState),
		CurrentState: currState,
		Metadata:     messages.MetadataMap{"seen": time.Now()},
	}
	if currStateWrapper.Verified {
		setupCurrStateCtx(currStateWrapper, cs)
		return csw.handleCurrentStateWrapper(cs, actorContext, currStateWrapper)
	}
	csw.Log.Errorw("invalid current state wrapper created internally!")

	return nil
}

func (csw *ConflictSetWorker) handleCurrentStateWrapper(cs *ConflictSet, context actor.Context, currWrapper *messages.CurrentStateWrapper) error {
	sp := cs.NewSpan("handleCurrentStateWrapper")
	defer sp.Finish()

	if !cs.active && (currWrapper.CurrentState.Signature.Height == currWrapper.NextHeight) {
		csw.Log.Debugw("msg.height equals msg.nextHeight, activating conflict set")
		sp.SetTag("activating", true)
		cs.active = true
	}

	currWrapperSpan := currWrapper.NewSpan("handleCurrentStateWrapper")
	defer currWrapper.StopTrace()
	defer currWrapperSpan.Finish()

	csw.Log.Debugw("handleCurrentStateWrapper", "internal", currWrapper.Internal, "verified", currWrapper.Verified)

	if currWrapper.Verified {
		if !cs.active {
			if cs.snoozedCommit != nil {
				return fmt.Errorf("received new commit with one already snoozed")
			}
			csw.Log.Debugw("snoozing commit")
			cs.snoozedCommit = currWrapper
			return nil
		}

		for _, t := range cs.transactions {
			transSpan := t.NewSpan("handleCurrentStateWrapper")
			transSpan.SetTag("done", true)
			if !bytes.Equal(t.Transaction.NewTip, currWrapper.CurrentState.Signature.NewTip) {
				transSpan.SetTag("error", true)
			}
			transSpan.Finish()
		}

		cs.done = true
		sp.SetTag("done", true)
		csw.Log.Debugw("conflict set is done")

		context.Send(csw.router, currWrapper)
		return nil
	}

	sp.SetTag("badSignature", true)
	sp.SetTag("error", true)
	csw.Log.Errorw("signature not verified AND SHOULD NEVER GET HERE")
	return nil
}

// returns a transaction with enough signatures or nil if none yet exist
func (csw *ConflictSetWorker) possiblyDone(cs *ConflictSet) *messages.TransactionWrapper {
	sp := cs.NewSpan("possiblyDone")
	defer sp.Finish()

	count := csw.notaryGroup.QuorumCount()
	csw.Log.Debugw("looking for a transaction with enough signatures", "quorum count", count)
	for tID, signature := range cs.signatures {
		if uint64(len(signature.Signers)) >= count {
			csw.Log.Debugw("found transaction with enough signatures", "id", tID)
			return cs.transactions[tID]
		} else {
			csw.Log.Debugw("count too low", "count", len(signature.Signers))
		}
	}
	csw.Log.Debugw("couldn't find any transaction with enough signatures")
	return nil
}

func (csw *ConflictSetWorker) deadlocked(cs *ConflictSet) bool {
	sp := cs.NewSpan("isDeadlocked")
	defer sp.Finish()

	unknownSigCount := len(csw.notaryGroup.Signers) - len(cs.hasSignedSomething)
	quorumAt := csw.notaryGroup.QuorumCount()
	if len(cs.hasSignedSomething) == 0 {
		return false
	}
	for _, sigs := range cs.signatures {
		if uint64(len(sigs.Signers)+unknownSigCount) >= quorumAt {
			return false
		}
	}

	return true
}

func sigToWrapper(sig *signatures.Signature, ng *types.NotaryGroup, self *types.Signer, isInternal bool) (*messages.SignatureWrapper, error) {
	signerMap := make(messages.SignerMap)

	allSigners := ng.AllSigners()
	for i, signer := range allSigners {
		cnt := sig.Signers[i]
		if cnt > 0 {
			signerMap[signer.ID] = signer
		}
	}

	conflictSetID := consensus.ConflictSetID(sig.ObjectId, sig.Height)

	committee, err := ng.RewardsCommittee([]byte(sig.NewTip), self)
	if err != nil {
		return nil, fmt.Errorf("error getting committee: %v", err)
	}

	return &messages.SignatureWrapper{
		Internal:         isInternal,
		ConflictSetID:    conflictSetID,
		RewardsCommittee: committee,
		Signers:          signerMap,
		Signature:        sig,
		Metadata:         messages.MetadataMap{"seen": time.Now()},
	}, nil
}

// this is a bit of custom tracing magic to give the currentStateWrapper a standard
// lifecycle of its own, but also link it up with this conflictSet
func setupCurrStateCtx(wrapper *messages.CurrentStateWrapper, cs *ConflictSet) {
	csSpan := opentracing.SpanFromContext(cs.GetContext())
	wrapperSpan := opentracing.StartSpan("currentStateWrapper", opentracing.FollowsFrom(csSpan.Context()))
	wrapperCtx := opentracing.ContextWithSpan(context.Background(), wrapperSpan)
	wrapper.SetContext(wrapperCtx)
}

func (csw *ConflictSetWorker) combineSignatures(a *messages.SignatureWrapper, b *messages.SignatureWrapper) (*messages.SignatureWrapper, error) {
	csw.Log.Debugw("combining signatures", "signersA", a.Signers, "signersB", b.Signers)
	newSignerCnt := make([]uint32, len(csw.notaryGroup.Signers))
	for i, cnt := range a.Signature.Signers {
		if right := b.Signature.Signers[i]; cnt > math.MaxUint32-right || right > math.MaxUint32-cnt {
			return nil, fmt.Errorf("error would overflow: %d %d", cnt, right)
		}
		newSignerCnt[i] = cnt + b.Signature.Signers[i]
	}

	newSignersMap := make(messages.SignerMap)
	for id, signer := range a.Signers {
		newSignersMap[id] = signer
	}

	for id, signer := range b.Signers {
		newSignersMap[id] = signer
	}

	aggregatedSig, err := bls.SumSignatures([][]byte{a.Signature.Signature, b.Signature.Signature})
	if err != nil {
		return nil, fmt.Errorf("error aggregating signatures: %v", err)
	}
	newSig := &signatures.Signature{
		TransactionId: a.Signature.TransactionId,
		ObjectId:      a.Signature.ObjectId,
		PreviousTip:   a.Signature.PreviousTip,
		Height:        a.Signature.Height,
		NewTip:        a.Signature.NewTip,
		View:          a.Signature.View,
		Cycle:         a.Signature.Cycle,
		Type:          a.Signature.Type,
		Signers:       newSignerCnt,
		Signature:     aggregatedSig,
	}

	return &messages.SignatureWrapper{
		Internal:         a.Internal || b.Internal,
		Signature:        newSig,
		Signers:          newSignersMap,
		RewardsCommittee: a.RewardsCommittee,
		// TODO: what about metadata?
	}, nil
}
