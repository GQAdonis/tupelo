package gossip

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-msgio"
	"github.com/multiformats/go-multihash"
	"github.com/quorumcontrol/chaintree/safewrap"
	"github.com/quorumcontrol/tupelo-go-sdk/p2p"

	"github.com/quorumcontrol/tupelo-go-sdk/gossip/types"
)

type snowballer struct {
	sync.RWMutex

	node     *Node
	snowball *Snowball
	height   uint64
	host     p2p.Node
	group    *types.NotaryGroup
	logger   logging.EventLogger
	cache    *lru.Cache

	earlyCommitter *earlyCommitter

	started bool
}

type snowballVoteResp struct {
	signerID   string
	checkpoint types.Checkpoint
}

func newSnowballer(n *Node, height uint64, snowball *Snowball) *snowballer {
	cache, err := lru.New(50)
	if err != nil {
		panic(err)
	}
	return &snowballer{
		node:           n,
		snowball:       snowball,
		host:           n.p2pNode,
		group:          n.notaryGroup,
		logger:         n.logger,
		cache:          cache,
		height:         height,
		earlyCommitter: newEarlyCommitter(),
	}
}

func (snb *snowballer) Started() bool {
	snb.RLock()
	defer snb.RUnlock()
	return snb.started
}

func (snb *snowballer) start(startCtx context.Context, done chan error) {
	sp := opentracing.StartSpan("gossip4.snowballer.start")
	defer sp.Finish()
	ctx := opentracing.ContextWithSpan(startCtx, sp)

	snb.Lock()
	snb.started = true
	snb.Unlock()

	signerCount := int(snb.group.Size())

	var earlyCommitted bool
	for !snb.snowball.Decided() && !earlyCommitted {
		snb.doTick(ctx)
		didEarly, checkpointID := snb.earlyCommitter.HasThreshold(signerCount, snb.snowball.alpha)
		if didEarly {
			snb.logger.Debugf("early commit on %s", checkpointID.String())
			earlyCommitted = didEarly
			checkpoint, ok := snb.cache.Get(checkpointID)
			if !ok {
				done <- errors.New("could not find checkpoint in the cache: should never happen")
				return
			}
			// TODO: this probably shouldn't reach into snowball here,
			// maybe it could move decided logic up to the snowballer here and others could reference this
			// rather than having the external system dig into the the snowball instance here too.
			snb.snowball.Prefer(&Vote{
				Checkpoint: checkpoint.(*types.Checkpoint),
			})
			snb.snowball.decided = true
		}
	}

	done <- nil
}

func (snb *snowballer) doTick(startCtx context.Context) {
	sp := opentracing.StartSpan("gossip4.snowballer.tick", opentracing.FollowsFrom(opentracing.SpanFromContext(startCtx).Context()))
	defer sp.Finish()
	ctx := opentracing.ContextWithSpan(startCtx, sp)

	sp.LogKV("height", snb.height)

	respChan := make(chan *snowballVoteResp, snb.snowball.k)
	wg := &sync.WaitGroup{}
	for i := 0; i < snb.snowball.k; i++ {
		wg.Add(1)
		go func() {
			snb.getOneRandomVote(ctx, wg, respChan)
		}()
	}
	wg.Wait()
	close(respChan)
	votes := make([]*Vote, snb.snowball.k)
	i := 0
	for resp := range respChan {
		checkpoint := resp.checkpoint
		vote := &Vote{
			Checkpoint: &checkpoint,
		}
		if len(checkpoint.AddBlockRequests) == 0 ||
			!snb.mempoolHasAllABRs(checkpoint.AddBlockRequests) ||
			snb.hasConflictingABRs(checkpoint.AddBlockRequests) {
			// nil out any votes that have ABRs we havne't heard of
			// or if they present conflicting ABRs in the same Checkpoint
			snb.logger.Debugf("nilling vote has all: %t, has conflicting: %t", snb.mempoolHasAllABRs(checkpoint.AddBlockRequests), snb.hasConflictingABRs(checkpoint.AddBlockRequests))
			vote.Nil()
		}
		votes[i] = vote
		// if the vote hasn't been nilled then we can add it to the early commiter
		if vote.ID() != ZeroVoteID {
			snb.logger.Debugf("received checkpoint: %v", checkpoint)
			snb.earlyCommitter.Vote(resp.signerID, resp.checkpoint.CID())
		}

		i++
	}

	if i < len(votes) {
		for j := i; j < len(votes); j++ {
			v := &Vote{}
			v.Nil()
			votes[j] = v
		}
	}

	votes = calculateTallies(votes)
	// snb.logger.Debugf("votes: %s", spew.Sdump(votes))

	snb.snowball.Tick(ctx, votes)
	// snb.logger.Debugf("counts: %v, beta: %d", snb.snowball.counts, snb.snowball.count)
}

func (snb *snowballer) getOneRandomVote(parentCtx context.Context, wg *sync.WaitGroup, respChan chan *snowballVoteResp) {
	sp, ctx := opentracing.StartSpanFromContext(parentCtx, "gossip4.snowballer.getOneRandomVote")
	defer sp.Finish()

	var signer *types.Signer
	var signerPeer string
	defer wg.Done()

	for signer == nil || signerPeer == snb.host.Identity() {
		signer = snb.group.GetRandomSigner()
		peerID, _ := p2p.PeerIDFromPublicKey(signer.DstKey)
		signerPeer = peerID.Pretty()
	}
	snb.logger.Debugf("snowballing with %s", signerPeer)
	sp.LogKV("signerPeer", signerPeer)

	s, err := snb.host.NewStream(ctx, signer.DstKey, gossipProtocol)
	if err != nil {
		sp.LogKV("error", err.Error())
		snb.logger.Warningf("error creating stream to %s: %v", signer.ID, err)
		if s != nil {
			s.Close()
		}
		return
	}
	defer s.Close()

	sw := &safewrap.SafeWrap{}
	wrapped := sw.WrapObject(snb.height)

	if err := s.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		sp.LogKV("error", err.Error())
		snb.logger.Errorf("error setting deadline: %v", err) // TODO: do we need to do anything about this error?
		return
	}

	writer := msgio.NewVarintWriter(s)
	err = writer.WriteMsg(wrapped.RawData())
	if err != nil {
		sp.LogKV("error", err.Error())
		snb.logger.Warningf("error writing to stream to %s: %v", signer.ID, err)
		return
	}

	reader := msgio.NewVarintReader(s)
	bits, err := reader.ReadMsg()
	if err != nil {
		sp.LogKV("error", err.Error())
		snb.logger.Warningf("error reading from stream to %s: %v", signer.ID, err)
		return
	}

	id, err := cidFromBits(bits)
	if err != nil {
		sp.LogKV("error", err.Error())
		snb.logger.Warningf("error creating bits to %s: %v", signer.ID, err)
		return
	}

	var checkpoint *types.Checkpoint

	blkInter, ok := snb.cache.Get(id)
	if ok {
		sp.LogKV("cacheHit", true)
		checkpoint = blkInter.(*types.Checkpoint)
	} else {
		sp.LogKV("cacheHit", false)
		blk := &types.Checkpoint{}
		err = cbornode.DecodeInto(bits, blk)
		if err != nil {
			sp.LogKV("error", err.Error())
			snb.logger.Warningf("error decoding from stream to %s: %v", signer.ID, err)
			return
		}
		checkpoint = blk
		snb.cache.Add(id, checkpoint)
	}

	respChan <- &snowballVoteResp{
		signerID:   signerPeer,
		checkpoint: *checkpoint,
	}
}

func cidFromBits(bits []byte) (cid.Cid, error) {
	hash, err := multihash.Sum(bits, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	return cid.NewCidV1(cid.DagCBOR, hash), nil
}

func (snb *snowballer) mempoolHasAllABRs(abrCIDs []cid.Cid) bool {
	hasAll := true
	for _, abrCID := range abrCIDs {
		ok := snb.node.mempool.Contains(abrCID)
		if !ok {
			snb.logger.Debugf("missing tx: %s", abrCID.String())
			snb.node.rootContext.Send(snb.node.syncerPid, abrCID)
			// the reason to not just return false here is that we want
			// to continue to loop over all the Txs in order to sync the ones we don't have
			hasAll = false
		}
	}
	return hasAll
}

// this checks to make sure the block coming in doesn't have any conflicting transactions
func (snb *snowballer) hasConflictingABRs(abrCIDs []cid.Cid) bool {
	csIds := make(map[mempoolConflictSetID]struct{})
	for _, abrCID := range abrCIDs {
		wrapper := snb.node.mempool.Get(abrCID)
		if wrapper == nil {
			snb.logger.Errorf("had a null transaction ( %s ) from a block in the mempool, this shouldn't happen", abrCID.String())
			return true
		}
		conflictSetID := toConflictSetID(wrapper.AddBlockRequest)
		_, ok := csIds[conflictSetID]
		if ok {
			return true
		}
		csIds[conflictSetID] = struct{}{}
	}
	return false
}

type snowballerDone struct {
	err error
	ctx context.Context
}
