package gossip4

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	cbornode "github.com/ipfs/go-ipld-cbor"

	"github.com/AsynkronIT/protoactor-go/actor"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-msgio"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-hamt-ipld"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/quorumcontrol/chaintree/nodestore"
	"github.com/quorumcontrol/chaintree/safewrap"
	"github.com/quorumcontrol/messages/v2/build/go/services"
	"github.com/quorumcontrol/tupelo-go-sdk/bls"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/types"
	"github.com/quorumcontrol/tupelo-go-sdk/p2p"
)

var genesis cid.Cid

const gossip4Protocol = "tupelo/v0.0.1"

const transactionTopic = "g4-transactions"

type mempool map[cid.Cid]*services.AddBlockRequest

func (m mempool) Keys() []cid.Cid {
	keys := make([]cid.Cid, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

type roundHolder map[uint64]*round

func init() {
	cbornode.RegisterCborType(services.AddBlockRequest{})
	sw := &safewrap.SafeWrap{}
	genesis = sw.WrapObject("genesis").Cid()
	if sw.Err != nil {
		panic(fmt.Errorf("error setting up genesis: %w", sw.Err))
	}
}

type snowballerDone struct {
	err error
	ctx context.Context
}

type snowballTicker struct {
	ctx context.Context
}

type Node struct {
	sync.RWMutex // it's weird to mix these synchronization primitives with actors, but it's expeditious at the moment - CODE CLEANUP please.

	name string

	p2pNode     p2p.Node
	signKey     *bls.SignKey
	notaryGroup *types.NotaryGroup
	dagStore    nodestore.DagStore
	hamtStore   *hamt.CborIpldStore
	pubsub      *pubsub.PubSub

	mempool  mempool
	inflight *lru.Cache

	logger logging.EventLogger

	rounds       roundHolder
	currentRound uint64

	snowballer *snowballer

	signerIndex int

	pid         *actor.PID
	snowballPid *actor.PID
	syncerPid   *actor.PID
	count       int
}

type NewNodeOptions struct {
	P2PNode        p2p.Node
	SignKey        *bls.SignKey
	NotaryGroup    *types.NotaryGroup
	DagStore       nodestore.DagStore
	Name           string
	PreviousRounds roundHolder
	CurrentRound   uint64
}

func NewNode(ctx context.Context, opts *NewNodeOptions) (*Node, error) {
	hamtStore := dagStoreToCborIpld(opts.DagStore)

	var signerIndex int
	for i, s := range opts.NotaryGroup.AllSigners() {
		if bytes.Equal(s.VerKey.Bytes(), opts.SignKey.MustVerKey().Bytes()) {
			signerIndex = i
			break
		}
	}

	if opts.Name == "" {
		opts.Name = fmt.Sprintf("node-%d", signerIndex)
	}

	logger := logging.Logger(opts.Name)

	logger.Debugf("signerIndex: %d", signerIndex)

	cache, err := lru.New(500)
	if err != nil {
		return nil, fmt.Errorf("error creating cache: %w", err)
	}

	if opts.PreviousRounds == nil {
		opts.PreviousRounds = make(roundHolder)
	}

	r, ok := opts.PreviousRounds[opts.CurrentRound]
	if !ok {
		r = newRound(opts.CurrentRound)
		opts.PreviousRounds[opts.CurrentRound] = r
	}

	n := &Node{
		name:         opts.Name,
		p2pNode:      opts.P2PNode,
		signKey:      opts.SignKey,
		notaryGroup:  opts.NotaryGroup,
		dagStore:     opts.DagStore,
		hamtStore:    hamtStore,
		rounds:       opts.PreviousRounds,
		currentRound: opts.CurrentRound,
		signerIndex:  signerIndex,
		logger:       logger,
		inflight:     cache,
		mempool:      make(mempool),
	}
	networkedSnowball := newSnowballer(n, r.height, r.snowball)
	n.snowballer = networkedSnowball
	return n, nil
}

// func loadNode(ctx context.Context, store *hamt.CborIpldStore, id cid.Cid) (*hamt.Node, error) {
// 	return hamt.LoadNode(ctx, store, id, hamt.UseTreeBitWidth(5))
// }

func (n *Node) setupSnowball(actorContext actor.Context) {
	snowballPid, err := actorContext.SpawnNamed(actor.PropsFromFunc(n.SnowBallReceive), "snowballProcess")
	if err != nil {
		panic(err)
	}
	n.snowballPid = snowballPid
	n.p2pNode.SetStreamHandler(gossip4Protocol, func(s network.Stream) {
		actor.EmptyRootContext.Send(snowballPid, s)
	})
}

func (n *Node) Start(ctx context.Context) error {
	pid, err := actor.EmptyRootContext.SpawnNamed(actor.PropsFromFunc(n.Receive), n.name)
	if err != nil {
		return fmt.Errorf("error starting actor: %w", err)
	}
	n.pid = pid

	n.logger.Debugf("pid: %s", pid.Id)

	go func() {
		<-ctx.Done()
		n.logger.Infof("node stopped having received: %d", n.count)
		actor.EmptyRootContext.Poison(n.pid)
	}()

	validator, err := newTransactionValidator(n.logger, n.notaryGroup, pid)
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

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if n.snowballPid != nil {
					actor.EmptyRootContext.Send(n.snowballPid, &snowballTicker{ctx: ctx})
				}
			}
		}
	}()

	n.logger.Debugf("node starting")

	return nil
}

func (n *Node) Receive(actorContext actor.Context) {
	switch msg := actorContext.Message().(type) {
	case *actor.Started:
		syncerPid, err := actorContext.SpawnNamed(actor.PropsFromProducer(func() actor.Actor {
			return &transactionGetter{
				nodeActor: actorContext.Self(),
				logger:    n.logger,
				store:     n.hamtStore,
			}
		}), "syncer")
		if err != nil {
			panic(err)
		}
		n.syncerPid = syncerPid
		n.setupSnowball(actorContext)
	case cid.Cid:
		actorContext.Forward(n.syncerPid)
	case *services.AddBlockRequest:
		n.handleAddBlockRequest(actorContext, msg)
	case *snowballerDone:
		n.handleSnowballerDone(msg)
	}
}

func (n *Node) handleSnowballerDone(msg *snowballerDone) {

	preferred := n.snowballer.snowball.Preferred()
	n.logger.Infof("round %d decided with err: %v: %s (len: %d)", n.currentRound, msg.err, preferred.ID(), len(preferred.Block.Transactions))
	n.logger.Debugf("round %d transactions %v", n.currentRound, preferred.Block.Transactions)
	// take all the transactions from the decided round, remove them from the mempool and apply them to the state
	// increase the currentRound and create a new Round in the roundHolder
	// state updating should be more robust here to make sure transactions don't stomp on each other and can probably happen in the background
	// probably need to recheck what's left in the mempool too for now we're ignoring all that as a PoC
	n.Lock() //TODO: WTF do I have locks in an actor?
	defer n.Unlock()
	n.logger.Infof("lock acquired for done: %d", preferred.Block.Height)

	round := newRound(n.currentRound + 1)
	n.rounds[n.currentRound+1] = round
	n.snowballer = newSnowballer(n, round.height, round.snowball)

	var rootNode *hamt.Node
	if n.currentRound == 0 {
		rootNode = hamt.NewNode(n.hamtStore, hamt.UseTreeBitWidth(5))
	} else {
		n.logger.Debugf("previous state for %d: %v", n.currentRound-1, n.rounds[n.currentRound-1].state)
		rootNode = n.rounds[n.currentRound-1].state.Copy()
	}
	n.logger.Debugf("current round: %d, node: %v", n.currentRound, rootNode)

	for _, txCID := range preferred.Block.Transactions {
		abr, ok := n.mempool[txCID]
		if !ok {
			n.logger.Errorf("I DO NOT HAVE THE TRANSACTION: %s", txCID.String())
			continue
		}
		err := rootNode.Set(msg.ctx, string(abr.ObjectId), txCID)
		if err != nil {
			panic(fmt.Errorf("error setting hamt: %w", err))
		}
		delete(n.mempool, txCID) //TODO: these should be shipped to a state updater
	}
	err := rootNode.Flush(msg.ctx)
	if err != nil {
		panic(fmt.Errorf("error flushing rootNode: %w", err))
	}
	n.logger.Debugf("setting round at %d to rootNode: %v", n.currentRound, rootNode)
	n.rounds[n.currentRound].state = rootNode
	n.logger.Debugf("after setting: %v", n.rounds[n.currentRound].state)
	n.currentRound++
}

func (n *Node) SnowBallReceive(actorContext actor.Context) {
	switch msg := actorContext.Message().(type) {
	case network.Stream:
		go func() {
			n.handleStream(msg)
		}()
	case *snowballTicker:
		n.RLock()
		defer n.RUnlock()
		if !n.snowballer.Started() && len(n.mempool) > 0 {
			go func() {
				n.RLock()
				n.logger.Debugf("starting snowballer and preferring %v", n.mempool.Keys())
				n.snowballer.snowball.Prefer(&Vote{
					Block: &Block{
						Height:       n.currentRound,
						Transactions: n.mempool.Keys(),
					},
				})
				n.RUnlock()
				done := make(chan error, 1)
				n.snowballer.start(msg.ctx, done)
				select {
				case <-msg.ctx.Done():
					return
				case err := <-done:
					actor.EmptyRootContext.Send(n.pid, &snowballerDone{err: err, ctx: msg.ctx})
				}
			}()
		}
	}
}

func (n *Node) handleStream(s network.Stream) {
	// n.logger.Debugf("handling stream from")
	s.SetWriteDeadline(time.Now().Add(2 * time.Second))
	s.SetReadDeadline(time.Now().Add(1 * time.Second))
	reader := msgio.NewVarintReader(s)
	writer := msgio.NewVarintWriter(s)

	bits, err := reader.ReadMsg()
	if err != nil {
		if err != io.EOF {
			_ = s.Reset()
		}
		n.logger.Warningf("error reading from incoming stream: %v", err)
		return
	}
	//TODO: I don't actually know what this does :)
	reader.ReleaseMsg(bits)

	var height uint64
	err = cbornode.DecodeInto(bits, &height)
	if err != nil {
		n.logger.Warningf("error decoding incoming height: %v", err)
		s.Close()
		return
	}
	// n.logger.Debugf("remote looking for height: %d", height)

	n.RLock()
	r, ok := n.rounds[height]
	n.RUnlock()

	response := Block{Height: height}
	if ok {
		// n.logger.Debugf("existing round %d", height)
		preferred := r.snowball.Preferred()
		if preferred != nil {
			// n.logger.Debugf("existing preferred; %v", preferred)
			response = *preferred.Block
		}
	}
	wrapped := response.Wrapped()

	err = writer.WriteMsg(wrapped.RawData())
	if err != nil {
		n.logger.Warningf("error writing: %v", err)
		s.Close()
		return
	}

	return
}

func (n *Node) handleAddBlockRequest(actorContext actor.Context, abr *services.AddBlockRequest) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	n.Lock()
	defer n.Unlock()
	n.count++

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
			// err = n.handlePostSave(ctx, actorContext, abr, current)
			// if err != nil {
			// 	n.logger.Errorf("error handling postSave: %v", err)
			// }
			return
		}
		n.storeAsInFlight(ctx, abr)
		return
	}

	// if this msg height is lower than current then just drop it
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
		// err = n.handlePostSave(ctx, actorContext, abr, current)
		// if err != nil {
		// 	n.logger.Errorf("error handling postSave: %v", err)
		// }
		return
	}

	if abr.Height > current.Height+1 {
		// this is in the future so just queue it up
		n.storeAsInFlight(ctx, abr)
		return
	}

	// TODO: handle byzantine case of msg.Height == current.Height
}

func (n *Node) storeAsInFlight(ctx context.Context, abr *services.AddBlockRequest) {
	n.logger.Infof("storing in inflight %s height: %d", string(abr.ObjectId), abr.Height)
	n.inflight.Add(inFlightID(abr.ObjectId, abr.Height), abr)
}

func inFlightID(objectID []byte, height uint64) string {
	return string(objectID) + strconv.FormatUint(height, 10)
}

func (n *Node) storeAbr(ctx context.Context, abr *services.AddBlockRequest) error {
	id, err := n.hamtStore.Put(ctx, abr)
	if err != nil {
		return fmt.Errorf("error putting abr: %w", err)
	}

	n.logger.Debugf("storing in mempool %s", id.String())
	n.mempool[id] = abr
	nextKey := inFlightID(abr.ObjectId, abr.Height+1)
	// if the next height Tx is here we can also queue that up
	next, ok := n.inflight.Get(nextKey)
	if ok {
		nextAbr := next.(*services.AddBlockRequest)
		if bytes.Equal(nextAbr.PreviousTip, abr.NewTip) {
			return n.storeAbr(ctx, nextAbr)
		}
		// This is commented out because I'm not sure exactly what we want to do here
		// if there is a big tree of Txs that aren't going to make it in, we don't necessarily
		// want to throw them away because we want the other nodes to know we have them.
		// // otherwise we can just throw that one away
		// n.inflight.Remove(nextKey)
	}

	return nil
}

func (n *Node) getCurrent(ctx context.Context, objectID string) (*services.AddBlockRequest, error) {
	if n.currentRound == 0 {
		return nil, nil // this is the genesis round: we, by definition, have no state
	}

	var abrCid cid.Cid

	lockedRound, ok := n.rounds[n.currentRound-1]
	if !ok {
		return nil, fmt.Errorf("we don't have the previous round")
	}

	n.logger.Debugf("previous round: %v", lockedRound)

	err := lockedRound.state.Find(ctx, objectID, &abrCid)
	if err != nil {
		if err == hamt.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting abrCID: %v", err)
	}

	abr := &services.AddBlockRequest{}

	if !abrCid.Equals(cid.Undef) {
		err = n.hamtStore.Get(ctx, abrCid, abr)
		if err != nil {
			return nil, fmt.Errorf("error getting abr: %v", err)
		}
	}

	return abr, nil
}
