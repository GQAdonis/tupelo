package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"math"
	"math/big"
	"sort"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/quorumcontrol/tupelo/bls"
	"github.com/quorumcontrol/tupelo/consensus"
)

type Signer struct {
	ID     string
	DstKey *ecdsa.PublicKey
	VerKey *bls.VerKey

	SignKey *bls.SignKey
	Actor   *actor.PID
}

func NewLocalSigner(dstKey *ecdsa.PublicKey, signKey *bls.SignKey) *Signer {
	pubKey := consensus.BlsKeyToPublicKey(signKey.MustVerKey())
	return &Signer{
		ID:      consensus.PublicKeyToAddr(&pubKey),
		SignKey: signKey,
		VerKey:  signKey.MustVerKey(),
		DstKey:  dstKey,
	}
}

func NewRemoteSigner(dstKey *ecdsa.PublicKey, verKey *bls.VerKey) *Signer {
	pubKey := consensus.BlsKeyToPublicKey(verKey)
	return &Signer{
		ID:     consensus.PublicKeyToAddr(&pubKey),
		VerKey: verKey,
		DstKey: dstKey,
	}
}

type NotaryGroup struct {
	Signers   map[string]*Signer
	sortedIds []string
}

func (ng *NotaryGroup) GetMajorityCount() int64 {
	required := int64(math.Ceil((2.0 * float64(len(ng.sortedIds))) / 3.0))
	if required == 0 {
		return 1
	}
	return required
}

func NewNotaryGroup() *NotaryGroup {
	return &NotaryGroup{
		Signers: make(map[string]*Signer),
	}
}

func (ng *NotaryGroup) AddSigner(signer *Signer) {
	ng.Signers[signer.ID] = signer
	ng.sortedIds = append(ng.sortedIds, signer.ID)
	sort.Strings(ng.sortedIds)
}

func (ng *NotaryGroup) AllSigners() []*Signer {
	signers := make([]*Signer, len(ng.sortedIds), len(ng.sortedIds))
	for i, id := range ng.sortedIds {
		signers[i] = ng.Signers[id]
	}
	return signers
}

func (ng *NotaryGroup) IndexOfSigner(signer *Signer) int {
	for i, s := range ng.sortedIds {
		if s == signer.ID {
			return i
		}
	}
	return -1
}

func (ng *NotaryGroup) GetRandomSyncer() *actor.PID {
	id := ng.sortedIds[randInt(len(ng.sortedIds)-1)]
	return ng.Signers[id].Actor
}

func randInt(max int) int {
	bigInt, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic("bad random")
	}
	return int(bigInt.Int64())
}

const minSyncNodesPerTransaction = 3

func (ng *NotaryGroup) RewardsCommittee(key []byte, excluding *Signer) ([]*Signer, error) {
	signerCount := float64(len(ng.sortedIds))
	logOfSigners := math.Log(signerCount)
	numberOfTargets := math.Min(signerCount-1, math.Floor(math.Max(logOfSigners, float64(minSyncNodesPerTransaction))))
	indexSpacing := signerCount / numberOfTargets
	moduloOffset := math.Mod(float64(bytesToUint64(key)), indexSpacing)

	targets := make([]*Signer, 0, int(numberOfTargets))
	i := 0
	for len(targets) < int(numberOfTargets) {
		targetIndex := int64(math.Floor(moduloOffset + (indexSpacing * float64(i))))
		targetID := ng.sortedIds[targetIndex]
		target := ng.Signers[targetID]
		// Make sure this node doesn't add itself as a target
		if target.ID == excluding.ID {
			continue
		}
		targets[i] = target
		i++
	}
	return targets, nil
}

func bytesToUint64(byteID []byte) uint64 {
	return binary.BigEndian.Uint64(byteID)
}