package gossip4

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	cbornode "github.com/ipfs/go-ipld-cbor"
	"github.com/multiformats/go-multihash"
	"github.com/opentracing/opentracing-go"

	"github.com/AsynkronIT/protoactor-go/actor"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-hamt-ipld"
	logging "github.com/ipfs/go-log"
	"github.com/quorumcontrol/chaintree/nodestore"
	"github.com/quorumcontrol/chaintree/safewrap"
	"github.com/quorumcontrol/messages/v2/build/go/services"
	"github.com/quorumcontrol/tupelo-go-sdk/bls"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/types"
	"github.com/quorumcontrol/tupelo-go-sdk/p2p"
)

var genesis cid.Cid

var biggest256BitNumber *big.Int
var difficultyThreshold *big.Int

const difficulty = 8 // checkpoint every X transactions or so

const transactionTopic = "g4-transactions"

func init() {
	cbornode.RegisterCborType(services.AddBlockRequest{})
	sw := &safewrap.SafeWrap{}
	genesis = sw.WrapObject("genesis").Cid()
	setupBiggest()
	setupThreshold()
}

func setupThreshold() {
	difficultyThreshold = big.NewInt(0)
	difficultyThreshold.Div(biggest256BitNumber, big.NewInt(difficulty))
}

func setupBiggest() {
	biggest256BitNumber = big.NewInt(0)
	biggest256BitNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // this is the largest 256bit number
}

type getInProgressCheckpoint struct{}

type Node struct {
	p2pNode     p2p.Node
	signKey     *bls.SignKey
	notaryGroup *types.NotaryGroup
	dagStore    nodestore.DagStore
	hamtStore   *hamt.CborIpldStore
	pubsub      *pubsub.PubSub

	inFlightAbrs map[string]*services.AddBlockRequest

	logger logging.EventLogger

	latestCheckpoint     *Checkpoint
	inprogressCheckpoint *Checkpoint

	signerIndex int

	pid *actor.PID
}

type NewNodeOptions struct {
	P2PNode          p2p.Node
	SignKey          *bls.SignKey
	NotaryGroup      *types.NotaryGroup
	DagStore         nodestore.DagStore
	Logger           logging.EventLogger
	latestCheckpoint cid.Cid
}

func NewNode(ctx context.Context, opts *NewNodeOptions) (*Node, error) {
	hamtStore := dagStoreToCborIpld(opts.DagStore)

	latest := &Checkpoint{}
	// if latestCheckpoint is undef then just make an empty hamt
	if opts.latestCheckpoint.Equals(cid.Undef) {
		n := hamt.NewNode(hamtStore, hamt.UseTreeBitWidth(5))

		latest = &Checkpoint{
			CurrentState: genesis,
			Previous:     genesis,
			Height:       0,
			node:         n,
		}
	} else {
		// otherwise pull up the CheckPoint
		err := hamtStore.Get(ctx, opts.latestCheckpoint, latest)
		if err != nil {
			return nil, fmt.Errorf("error getting latest: %v", err)
		}
		n, err := loadNode(ctx, hamtStore, latest.CurrentState)
		if err != nil {
			return nil, fmt.Errorf("error loading node: %v", err)
		}
		latest.node = n
	}

	// then clone for the inprogress
	inProgress := latest.Copy()
	inProgress.CurrentState = cid.Undef

	var signerIndex int
	for i, s := range opts.NotaryGroup.AllSigners() {
		if bytes.Equal(s.VerKey.Bytes(), opts.SignKey.MustVerKey().Bytes()) {
			signerIndex = i
			break
		}
	}

	logger := logging.Logger(fmt.Sprintf("node-%d", signerIndex))

	logger.Debugf("signerIndex: %d", signerIndex)

	return &Node{
		p2pNode:              opts.P2PNode,
		signKey:              opts.SignKey,
		notaryGroup:          opts.NotaryGroup,
		dagStore:             opts.DagStore,
		hamtStore:            hamtStore,
		latestCheckpoint:     latest,
		inprogressCheckpoint: inProgress,
		inFlightAbrs:         make(map[string]*services.AddBlockRequest),
		signerIndex:          signerIndex,
		logger:               logger,
	}, nil
}

func loadNode(ctx context.Context, store *hamt.CborIpldStore, id cid.Cid) (*hamt.Node, error) {
	return hamt.LoadNode(ctx, store, id, hamt.UseTreeBitWidth(5))
}

func (n *Node) Start(ctx context.Context) error {
	pid := actor.EmptyRootContext.Spawn(actor.PropsFromFunc(n.Receive))
	n.pid = pid
	go func() {
		<-ctx.Done()
		n.logger.Debugf("node stopped")
		actor.EmptyRootContext.Poison(pid)
	}()

	validator, err := newTransactionValidator(n.logger, n.notaryGroup, n.pid)
	if err != nil {
		return fmt.Errorf("error setting up: %v", err)
	}

	n.pubsub = n.p2pNode.GetPubSub()

	err = n.pubsub.RegisterTopicValidator(transactionTopic, validator.validate)
	if err != nil {
		return fmt.Errorf("error registering topic validator: %v", err)
	}

	sub, err := n.pubsub.Subscribe(transactionTopic)
	if err != nil {
		return fmt.Errorf("error subscribing %v", err)
	}

	// don't do anything with these messages because we actually get them
	// fully decoded in the actor spun up above
	go func() {
		for {
			_, err := sub.Next(ctx)
			if err != nil {
				n.logger.Warningf("error getting sub message: %v", err)
				return
			}
		}
	}()

	n.logger.Debugf("node starting with difficulty: %0256b", difficultyThreshold)

	return nil
}

func (n *Node) Receive(actorContext actor.Context) {
	switch msg := actorContext.Message().(type) {
	case *services.AddBlockRequest:
		n.handleAddBlockRequest(actorContext, msg)
	case *getInProgressCheckpoint:
		n.logger.Debugf("returning inprogress checkpoint of height: %d", n.inprogressCheckpoint.Height)
		actorContext.Respond(n.inprogressCheckpoint)
	}
}

func (n *Node) commitCheckpoint(actorContext actor.Context, cp *Checkpoint) {
	n.logger.Debugf("committing a checkpoint %d", cp.Height)

	if cp.node == nil {
		node, err := loadNode(context.TODO(), n.hamtStore, cp.CurrentState)
		if err != nil {
			n.logger.Errorf("error getting node: %v", err)
		}
		cp.node = node
	}

	n.logger.Debugf("resetting the inprogress checkpoint")
	n.latestCheckpoint = cp
	n.inprogressCheckpoint = cp.Copy()
	n.inprogressCheckpoint.Height = cp.Height + 1
	n.inprogressCheckpoint.Previous = cp.CurrentState
	n.inprogressCheckpoint.PreviousSignature = cp.Signature
	n.logger.Infof("committing %d %s", cp.Height, cp.CurrentState.String())

	n.logger.Debugf("deleting inflight transactions that have been committed (height now: %d)", n.inprogressCheckpoint.Height)
	for _, txID := range cp.Transactions {
		n.logger.Infof("committed %s", txID.String())
		abr, ok := n.inFlightAbrs[txID.String()]
		if ok {
			delete(n.inFlightAbrs, txID.String())
			delete(n.inFlightAbrs, heightKey(string(abr.ObjectId), abr.Height))
		}
	}

	for k, abr := range n.inFlightAbrs {
		if strings.HasPrefix(k, "FH") {
			continue // don't do anything with the height ones
		}
		n.logger.Debugf("requeing tx: %s", k)
		actorContext.Send(actorContext.Self(), abr)
	}
}

func (n *Node) handleAddBlockRequest(actorContext actor.Context, abr *services.AddBlockRequest) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	sp := opentracing.StartSpan("add-block-request-gossip4-node")
	defer sp.Finish()

	n.logger.Debugf("handling message: ObjectId: %s, Height: %d", abr.ObjectId, abr.Height)
	// look into the hamt, and get the current ABR
	// if this is the next height then replace it
	// if this is older drop it
	// if this is in the future, keep it in flight

	current, err := n.getCurrent(ctx, string(abr.ObjectId))
	if err != nil {
		n.logger.Errorf("error getting current: %v", err)
		return
	}

	// if the current is nil, then this ABR is acceptable if its height is 0
	if current == nil {
		if abr.Height == 0 {
			err = n.storeAbr(ctx, abr)
			if err != nil {
				n.logger.Errorf("error getting current: %v", err)
				return
			}
			err = n.handlePostSave(ctx, actorContext, abr, current)
			if err != nil {
				n.logger.Errorf("error handling postSave: %v", err)
			}
			return
		}
		n.storeAsInFlight(ctx, abr)
		return
	}

	// if this msg height is lower than curreent then just drop it
	if current.Height > abr.Height {
		return
	}

	// Is this the next height ABR?
	if current.Height+1 == abr.Height {
		// then this is the next height, let's save it if the tips match

		if !bytes.Equal(current.PreviousTip, abr.PreviousTip) {
			n.logger.Warningf("tips did not match")
			return
		}

		err = n.storeAbr(ctx, abr)
		if err != nil {
			n.logger.Errorf("error getting current: %v", err)
			return
		}
		err = n.handlePostSave(ctx, actorContext, abr, current)
		if err != nil {
			n.logger.Errorf("error handling postSave: %v", err)
		}
		return
	}

	if abr.Height > current.Height+1 {
		// this is in the future so just queue it up
		n.storeAsInFlight(ctx, abr)
		return
	}

	// TODO: handle byzantine case of msg.Height == current.Height
}

func (n *Node) handlePostSave(ctx context.Context, actorContext actor.Context, saved *services.AddBlockRequest, overwritten *services.AddBlockRequest) error {
	// Is this a checkpoint?
	id, err := n.hamtStore.Put(ctx, n.inprogressCheckpoint.node)
	if err != nil {
		return fmt.Errorf("error putting checkpoint: %v", err)
	}

	n.inprogressCheckpoint.CurrentState = id

	if n.isCidDifficultEnough(id) {
		err := n.inprogressCheckpoint.node.Flush(ctx)
		if err != nil {
			return fmt.Errorf("error flushing: %v", err)
		}
		n.commitCheckpoint(actorContext, n.inprogressCheckpoint)
	}

	return nil
}

// is the digest of the CID
func (n *Node) isCidDifficultEnough(id cid.Cid) bool {
	decoded, _ := multihash.Decode(id.Hash()) // can never error
	hsh := decoded.Digest
	num := big.NewInt(0)
	num.SetBytes(hsh)
	if cmpRet := num.Cmp(difficultyThreshold); cmpRet > 0 { // if the difficultyThreshold is bigger than the digest, it's not difficult enough
		return false
	}
	return true
}

func heightKey(objectId string, height uint64) string {
	return "FH" + objectId + "-" + strconv.FormatUint(height, 10)
}

func (n *Node) storeAsInFlight(ctx context.Context, abr *services.AddBlockRequest) {
	n.inFlightAbrs[heightKey(string(abr.ObjectId), abr.Height)] = abr
}

func (n *Node) storeAbr(ctx context.Context, abr *services.AddBlockRequest) error {
	id, err := n.hamtStore.Put(ctx, abr)
	if err != nil {
		n.logger.Errorf("error putting abr: %v", err)
		return err
	}
	n.logger.Infof("saved abr with CID %s ", id.String())
	n.logger.Debugf("(objectId: %s, height: %d) in progress height: %d", abr.ObjectId, abr.Height, n.inprogressCheckpoint.Height)
	err = n.inprogressCheckpoint.node.Set(ctx, string(abr.ObjectId), id)
	if err != nil {
		n.logger.Errorf("error setting hamt: %v", err)
		return err
	}
	n.inFlightAbrs[id.String()] = abr // this is useful because we can just easily confirm we have this transaction
	n.inprogressCheckpoint.Transactions = append(n.inprogressCheckpoint.Transactions, id)
	return nil
}

func (n *Node) getCurrent(ctx context.Context, objectID string) (*services.AddBlockRequest, error) {
	var abrCid cid.Cid
	var abr *services.AddBlockRequest
	err := n.inprogressCheckpoint.node.Find(ctx, objectID, &abrCid)
	if err != nil {
		if err == hamt.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting abrCID: %v", err)
	}

	if !abrCid.Equals(cid.Undef) {
		err = n.hamtStore.Get(ctx, abrCid, abr)
		if err != nil {
			return nil, fmt.Errorf("error getting abr: %v", err)
		}
	}

	return abr, nil
}

func (n *Node) PID() *actor.PID {
	return n.pid
}

/**
when a transaction comes in make sure it's internally consistent and the height is higher than current
return true to the pubsub validator and pipe the tx to the node.

At the node do the normal thing where if it's the next height, then it can go into the hamt as accepted.

Put it in the transaction hamt, update the current state hamt. If the difficulty level has hit, publish out the CheckPoint

On receiving a CheckPoint...
	* confirm it meets the difficulty level
	* confirm it's a height higher than where we are at
	* confirm all transactions are valid
	* sign it
	* publish out the updated signature

*/