package gossip

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/quorumcontrol/qc3/bls"
	"github.com/quorumcontrol/qc3/consensus"
	"github.com/quorumcontrol/qc3/network"
	"github.com/quorumcontrol/qc3/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lastAccepted []byte
var lastRejected []byte

type testSet struct {
	SignKeys          []*bls.SignKey
	VerKeys           []*bls.VerKey
	EcdsaKeys         []*ecdsa.PrivateKey
	PubKeys           []consensus.PublicKey
	SignKeysByAddress map[string]*bls.SignKey
}

// This is the simplest possible state handler that just always returns the transaction as the nextState
func simpleHandler(_ context.Context, _ *consensus.Group, _, transaction, _ []byte) (nextState []byte, accepted bool, err error) {
	if bytes.HasPrefix(transaction, []byte("reject")) {
		log.Debug("rejecting transaction")
		return []byte("bad state no use"), false, nil
	}
	return transaction, true, nil
}

func simpleAcceptance(_ context.Context, _ *consensus.Group, _, transaction, _ []byte) (err error) {
	log.Debug("simpleAcceptance called")
	lastAccepted = transaction
	return nil
}

func simpleRejecter(ctx context.Context, group *consensus.Group, objectId, transaction, currentState []byte) (err error) {
	log.Debug("simpleRejecter called")
	lastRejected = transaction
	return nil
}

type InMemoryHandlerSystem struct {
	Handlers          map[string]*InMemoryHandler
	ArtificialLatency int
}

func NewInMemoryHandlerSystem(latency int) *InMemoryHandlerSystem {
	return &InMemoryHandlerSystem{
		Handlers:          make(map[string]*InMemoryHandler),
		ArtificialLatency: latency,
	}
}

func (imhs *InMemoryHandlerSystem) NewHandler(dst *ecdsa.PublicKey) *InMemoryHandler {
	imh := &InMemoryHandler{
		Id:     crypto.PubkeyToAddress(*dst).String(),
		System: imhs,
	}
	imhs.Handlers[imh.Id] = imh
	return imh
}

type InMemoryHandler struct {
	Id      string
	System  *InMemoryHandlerSystem
	Mapping map[string]network.HandlerFunc
}

func (imh *InMemoryHandler) Push(dst *ecdsa.PublicKey, req *network.Request) error {
	go func(dst *ecdsa.PublicKey, req *network.Request) {
		log.Trace("DoRequest go func started", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id, "elapsed")
		handler, ok := imh.System.Handlers[crypto.PubkeyToAddress(*dst).String()]
		if !ok {
			log.Error("could not find handler")
			panic("could not find handler")
		}
		internalRespChan := make(network.ResponseChan, 1)
		defer close(internalRespChan)
		err := handler.Mapping[req.Type](context.Background(), *req, internalRespChan)
		if err != nil {
			log.Error("error handling request", "err", err)
		}

		log.Trace("DoRequest func executed", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id)
		<-internalRespChan
	}(dst, req)

	return nil
}

func (imh *InMemoryHandler) DoRequest(dst *ecdsa.PublicKey, req *network.Request) (chan *network.Response, error) {
	respChan := make(chan *network.Response, 1)
	start := time.Now()

	log.Trace("DoRequest to", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id, "start", start)
	go func(dst *ecdsa.PublicKey, req *network.Request) {
		log.Trace("DoRequest go func started", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id, "elapsed", time.Now().Sub(start))
		handler, ok := imh.System.Handlers[crypto.PubkeyToAddress(*dst).String()]
		if !ok {
			log.Error("could not find handler")
			panic("could not find handler")
		}
		internalRespChan := make(network.ResponseChan, 1)
		defer close(respChan)

		err := handler.Mapping[req.Type](context.Background(), *req, internalRespChan)
		if err != nil {
			log.Error("error handling request", "err", err)
		}

		log.Trace("DoRequest func executed", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id, "elapsed", time.Now().Sub(start))
		<-time.After(time.Duration(imh.System.ArtificialLatency) * time.Millisecond)
		log.Trace("DoRequest responding", "dst", crypto.PubkeyToAddress(*dst).String(), "uuid", req.Id, "elapsed", time.Now().Sub(start))
		respChan <- <-internalRespChan
	}(dst, req)

	return respChan, nil
}

func (imh *InMemoryHandler) AssignHandler(requestType string, handlerFunc network.HandlerFunc) error {
	if imh.Mapping == nil {
		imh.Mapping = make(map[string]network.HandlerFunc)
	}

	imh.Mapping[requestType] = handlerFunc
	return nil
}

func (imh *InMemoryHandler) Start() {}
func (imh *InMemoryHandler) Stop()  {}

func newTestSet(t *testing.T, size int) *testSet {
	signKeys := blsKeys(size)
	verKeys := make([]*bls.VerKey, len(signKeys))
	pubKeys := make([]consensus.PublicKey, len(signKeys))
	ecdsaKeys := make([]*ecdsa.PrivateKey, len(signKeys))
	signKeysByAddress := make(map[string]*bls.SignKey)
	for i, signKey := range signKeys {
		ecdsaKey, _ := crypto.GenerateKey()
		verKeys[i] = signKey.MustVerKey()
		pubKeys[i] = consensus.BlsKeyToPublicKey(verKeys[i])
		ecdsaKeys[i] = ecdsaKey
		signKeysByAddress[consensus.BlsVerKeyToAddress(verKeys[i].Bytes()).String()] = signKey

	}

	return &testSet{
		SignKeys:          signKeys,
		VerKeys:           verKeys,
		PubKeys:           pubKeys,
		EcdsaKeys:         ecdsaKeys,
		SignKeysByAddress: signKeysByAddress,
	}
}

func groupFromTestSet(t *testing.T, set *testSet) *consensus.Group {
	members := make([]*consensus.RemoteNode, len(set.SignKeys))
	for i := range set.SignKeys {
		rn := consensus.NewRemoteNode(consensus.BlsKeyToPublicKey(set.VerKeys[i]), consensus.EcdsaToPublicKey(&set.EcdsaKeys[i].PublicKey))
		members[i] = rn
	}

	return consensus.NewGroup(members)
}

func blsKeys(size int) []*bls.SignKey {
	keys := make([]*bls.SignKey, size)
	for i := 0; i < size; i++ {
		keys[i] = bls.MustNewSignKey()
	}
	return keys
}

func generateTestGossipGroup(t *testing.T, size int, latency int) []*Gossiper {
	ts := newTestSet(t, size)
	group := groupFromTestSet(t, ts)

	system := NewInMemoryHandlerSystem(latency)

	gossipers := make([]*Gossiper, size)

	for i := 0; i < size; i++ {
		stor := storage.NewMemStorage()

		member := group.SortedMembers[i]

		gossiper := &Gossiper{
			SignKey:         ts.SignKeysByAddress[member.Id],
			Storage:         stor,
			StateHandler:    simpleHandler,
			AcceptedHandler: simpleAcceptance,
			RejectedHandler: simpleRejecter,
			Group:           group,
			MessageHandler:  system.NewHandler(crypto.ToECDSAPub(member.DstKey.PublicKey)),
		}

		gossiper.Initialize()
		gossipers[i] = gossiper
	}

	return gossipers
}

func TestTest(t *testing.T) {
	gossipers := generateTestGossipGroup(t, 2, 0)

	key := gossipers[0].SignKey
	msg := []byte("hi")

	sig, err := consensus.BlsSignBytes(crypto.Keccak256(msg), key)
	require.Nil(t, err)

	pubKey := consensus.BlsKeyToPublicKey(key.MustVerKey())

	verified, err := consensus.Verify(crypto.Keccak256(msg), *sig, pubKey)
	require.Nil(t, err)
	require.True(t, verified)

	sigMap := consensus.SignatureMap{gossipers[0].ID: *sig}

	groupSig, err := gossipers[0].Group.CombineSignatures(sigMap)
	require.Nil(t, err)

	verified, err = gossipers[0].Group.VerifyAvailableSignatures(crypto.Keccak256(msg), groupSig)
	require.Nil(t, err)
	assert.True(t, verified)
}

func TestGossiper_HandleGossip(t *testing.T) {
	//log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(log.LvlDebug), log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	gossipers := generateTestGossipGroup(t, 3, 0)

	type testCase struct {
		description string
		message     *GossipMessage
		shouldErr   bool
	}

	type createTestCase func() *testCase

	for _, testCaseCreator := range []createTestCase{
		func() *testCase {
			return &testCase{
				description: "prepare messsage no sigs",
				message: &GossipMessage{
					ObjectID:    []byte("obj"),
					PreviousTip: nil,
					Transaction: []byte("trans"),
					Phase:       phasePrepare,
					Round:       gossipers[0].RoundAt(time.Now()),
				},
			}
		},
		func() *testCase {
			return &testCase{
				description: "tentativeCommit messsage no sigs",
				shouldErr:   true,
				message: &GossipMessage{
					ObjectID:    []byte("obj"),
					PreviousTip: nil,
					Transaction: []byte("trans"),
					Phase:       phaseTentativeCommit,
					Round:       gossipers[0].RoundAt(time.Now()),
				},
			}
		},
	} {
		tc := testCaseCreator()
		err := gossipers[0].HandleGossip(context.TODO(), tc.message)
		if tc.shouldErr {
			require.NotNil(t, err, tc.description)
		} else {
			require.Nil(t, err, tc.description)
		}
	}
}
func TestGossiper_RoundHandlers(t *testing.T) {
	gossipers := generateTestGossipGroup(t, 1, 0)
	gossipers[0].RoundLength = 1
	var lastCalled int64
	currRound := gossipers[0].RoundAt(time.Now())
	gossipers[0].AddRoundHandler(func(_ context.Context, round int64) {
		lastCalled = round
	})
	gossipers[0].Start()
	defer gossipers[0].Stop()
	time.Sleep(time.Duration(gossipers[0].RoundLength) * time.Second)
	assert.Equal(t, lastCalled, currRound+1)
}
