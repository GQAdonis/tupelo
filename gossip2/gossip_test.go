package gossip2

import (
	"context"
	"crypto/ecdsa"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-log"
	"github.com/quorumcontrol/chaintree/nodestore"
	"github.com/quorumcontrol/qc3/bls"
	"github.com/quorumcontrol/qc3/consensus"
	"github.com/quorumcontrol/qc3/p2p"
	"github.com/quorumcontrol/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBootstrapHost(ctx context.Context, t *testing.T) *p2p.Host {
	key, err := crypto.GenerateKey()
	require.Nil(t, err)

	host, err := p2p.NewHost(ctx, key, p2p.GetRandomUnusedPort())

	require.Nil(t, err)
	require.NotNil(t, host)

	return host
}

func bootstrapAddresses(bootstrapHost *p2p.Host) []string {
	addresses := bootstrapHost.Addresses()
	for _, addr := range addresses {
		addrStr := addr.String()
		if strings.Contains(addrStr, "127.0.0.1") {
			return []string{addrStr}
		}
	}
	return nil
}

type testSet struct {
	SignKeys          []*bls.SignKey
	VerKeys           []*bls.VerKey
	EcdsaKeys         []*ecdsa.PrivateKey
	PubKeys           []consensus.PublicKey
	SignKeysByAddress map[string]*bls.SignKey
}

func newTestSet(t *testing.T, size int) *testSet {
	signKeys := blsKeys(size)
	verKeys := make([]*bls.VerKey, len(signKeys))
	pubKeys := make([]consensus.PublicKey, len(signKeys))
	ecdsaKeys := make([]*ecdsa.PrivateKey, len(signKeys))
	signKeysByAddress := make(map[string]*bls.SignKey)
	for i, signKey := range signKeys {
		ecdsaKey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("error generating key: %v", err)
		}
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

func groupFromTestSet(t *testing.T, set *testSet) *consensus.NotaryGroup {
	members := make([]*consensus.RemoteNode, len(set.SignKeys))
	for i := range set.SignKeys {
		rn := consensus.NewRemoteNode(consensus.BlsKeyToPublicKey(set.VerKeys[i]), consensus.EcdsaToPublicKey(&set.EcdsaKeys[i].PublicKey))
		members[i] = rn
	}

	nodeStore := nodestore.NewStorageBasedStore(storage.NewMemStorage())
	group := consensus.NewNotaryGroup("notarygroupid", nodeStore)
	err := group.CreateGenesisState(group.RoundAt(time.Now()), members...)
	require.Nil(t, err)
	return group
}

func blsKeys(size int) []*bls.SignKey {
	keys := make([]*bls.SignKey, size)
	for i := 0; i < size; i++ {
		keys[i] = bls.MustNewSignKey()
	}
	return keys
}

func randBytes(length int) []byte {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("couldn't generate random bytes")
	}
	return b
}

func TestExists(t *testing.T) {
	path := "existsTest"
	os.RemoveAll(path)
	os.MkdirAll(path, 0755)
	defer os.RemoveAll(path)

	storage := NewBadgerStorage(path)
	key := []byte("abcx")
	err := storage.Set(key, []byte("a"))
	require.Nil(t, err)
	exists, err := storage.Exists(key)
	require.Nil(t, err)
	assert.True(t, exists)
}

func TestGossip(t *testing.T) {
	logging.SetLogLevel("gossip", "ERROR")
	groupSize := 20
	ts := newTestSet(t, groupSize)
	group := groupFromTestSet(t, ts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bootstrap := newBootstrapHost(ctx, t)

	gossipNodes := make([]*GossipNode, groupSize)

	for i := 0; i < groupSize; i++ {
		host, err := p2p.NewHost(ctx, ts.EcdsaKeys[i], p2p.GetRandomUnusedPort())
		require.Nil(t, err)
		host.Bootstrap(bootstrapAddresses(bootstrap))
		path := "test/badger/" + strconv.Itoa(i)
		os.RemoveAll(path)
		os.MkdirAll(path, 0755)
		defer func() {
			os.RemoveAll(path)
		}()
		storage := NewBadgerStorage(path)
		gossipNodes[i] = NewGossipNode(ts.EcdsaKeys[i], host, storage)
		gossipNodes[i].Group = group
		gossipNodes[i].SignKey = ts.SignKeys[i]
	}

	transaction1 := Transaction{
		ObjectID:    []byte("himynameisalongobjectidthatwillhavemorethan64bits"),
		PreviousTip: []byte(""),
		NewTip:      []byte("zdpuAs5LQAGsXbGTF3DbfGVkRw4sWJd4MzbbigtJ4zE6NNJrr"),
		Payload:     []byte("thisisthepayload"),
	}
	log.Debugf("gossipNode0 is %s", gossipNodes[0].ID())

	// transaction2 := Transaction{
	// 	ObjectID:    []byte("DIFFERENTOBJhimynameisalongobjectidthatwillhavemorethan64bits"),
	// 	PreviousTip: []byte(""),
	// 	NewTip:      []byte("zdpuAx6tV9jpLEwhvGB8bdYjUvkomzdHf7ze6ckdNqJur7JBr"),
	// 	Payload:     []byte("thisisthepayload"),
	// }
	// var transaction2ID []byte
	// for i := 1; i < len(gossipNodes)-1; i++ {
	// 	transaction2ID, err := gossipNodes[i].InitiateTransaction(transaction2)
	// 	require.Nil(t, err)
	// }

	for i := 0; i < groupSize; i++ {
		defer gossipNodes[i].Stop()
		go gossipNodes[i].Start()
	}

	// This bit of commented out code will run the CPU profiler
	// f, ferr := os.Create("../gossip.prof")
	// if ferr != nil {
	// 	t.Fatal(ferr)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	for i := 0; i < 20; i++ {
		_, err := gossipNodes[rand.Intn(len(gossipNodes))].InitiateTransaction(Transaction{
			ObjectID:    randBytes(32),
			PreviousTip: []byte(""),
			NewTip:      randBytes(49),
			Payload:     randBytes(rand.Intn(400) + 100),
		})
		if err != nil {
			t.Fatalf("error sending transaction: %v", err)
		}
	}

	start := time.Now()
	_, err := gossipNodes[0].InitiateTransaction(transaction1)
	require.Nil(t, err)

	var stop time.Time
	for {
		if (time.Now().Sub(start)) > (60 * time.Second) {
			t.Fatal("timed out looking for done function")
			break
		}
		exists, err := gossipNodes[0].Storage.Exists(transaction1.ToConflictSet().DoneID())
		require.Nil(t, err)
		if exists {
			stop = time.Now()
			break
		}
	}

	// fmt.Printf("Confirmation took %f seconds\n", stop.Sub(start).Seconds())
	t.Logf("Confirmation took %f seconds\n", stop.Sub(start).Seconds())
	assert.True(t, stop.Sub(start) < 10*time.Second)

	// time.Sleep(2 * time.Second)
	// for i := 0; i < 1000; i++ {
	// 	if (i%10) == 0 && i < 950 {
	// 		log.Debugf("gossipNode0 is %s", gossipNodes[0].ID())
	// 		_, err := gossipNodes[rand.Intn(len(gossipNodes))].InitiateTransaction(Transaction{
	// 			ObjectID:    randBytes(32),
	// 			PreviousTip: []byte(""),
	// 			NewTip:      randBytes(49),
	// 			Payload:     randBytes(rand.Intn(400) + 100),
	// 		})
	// 		require.Nil(t, err)
	// 	}

	// 	err = gossipNodes[rand.Intn(len(gossipNodes))].DoSync()
	// 	require.Nil(t, err)
	// 	log.Debugf("")
	// 	log.Debugf("+++++++++++++++++++++++++++")
	// 	log.Debugf("ITEREATION %v done", i)
	// 	log.Debugf("+++++++++++++++++++++++++++")
	// }
	// time.Sleep(3 * time.Second)

	for i := 0; i < 1; i++ {
		exists, err := gossipNodes[i].Storage.Exists(transaction1.ToConflictSet().DoneID())
		require.Nil(t, err)
		assert.True(t, exists)

		// val, err = gossipNodes[i].Storage.Get(transaction2ID)
		// require.Nil(t, err)
		// encodedTransaction2, err := transaction2.MarshalMsg(nil)
		// require.Nil(t, err)
		// assert.Equal(t, val, encodedTransaction2)
	}

	csq := ConflictSetQuery{Key: transaction1.ToConflictSet().ID()}
	csqBytes, err := csq.MarshalMsg(nil)
	require.Nil(t, err)

	csqrBytes, err := gossipNodes[0].Host.SendAndReceive(&gossipNodes[1].Key.PublicKey, IsDoneProtocol, csqBytes)
	csqr := ConflictSetQueryResponse{}
	_, err = csqr.UnmarshalMsg(csqrBytes)
	require.Nil(t, err)
	assert.True(t, csqr.Done)

	var totalIn int64
	var totalOut int64
	var totalReceiveSync int64
	var totalAttemptSync int64
	for _, gn := range gossipNodes {
		totalIn += gn.Host.Reporter.GetBandwidthTotals().TotalIn
		totalOut += gn.Host.Reporter.GetBandwidthTotals().TotalOut
		totalReceiveSync += int64(atomic.LoadUint64(&gn.debugReceiveSync))
		totalAttemptSync += int64(atomic.LoadUint64(&gn.debugAttemptSync))
	}
	t.Logf(
		`Bandwidth In: %d, Bandwidth Out: %d,
Average In %d, Average Out %d
Total Receive Syncs: %d, Average Receive Sync: %d
Total Attempted Syncs: %d, Average Attepted Syncs: %d`,
		totalIn,
		totalOut,
		totalIn/int64(len(gossipNodes)),
		totalOut/int64(len(gossipNodes)),
		totalReceiveSync,
		totalReceiveSync/int64(len(gossipNodes)),
		totalAttemptSync,
		totalAttemptSync/int64(len(gossipNodes)),
	)

}
