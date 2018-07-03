package consensus

import (
	"testing"

	"github.com/quorumcontrol/qc3/bls"
	"github.com/stretchr/testify/assert"
)

type testSet struct {
	SignKeys []*bls.SignKey
	VerKeys  []*bls.VerKey
	PubKeys  []PublicKey
}

func newTestSet(t *testing.T) *testSet {
	signKeys := blsKeys(5)
	verKeys := make([]*bls.VerKey, len(signKeys))
	pubKeys := make([]PublicKey, len(signKeys))
	for i, signKey := range signKeys {
		verKeys[i] = signKey.MustVerKey()
		pubKeys[i] = BlsKeyToPublicKey(verKeys[i])
	}

	return &testSet{
		SignKeys: signKeys,
		VerKeys:  verKeys,
		PubKeys:  pubKeys,
	}
}

func blsKeys(size int) []*bls.SignKey {
	keys := make([]*bls.SignKey, size)
	for i := 0; i < size; i++ {
		keys[i] = bls.MustNewSignKey()
	}
	return keys
}

func TestGroupFromPublicKeys(t *testing.T) {
	ts := newTestSet(t)
	g := GroupFromPublicKeys(ts.PubKeys)
	assert.IsType(t, &Group{}, g)
}

func TestGroup_CombineSignatures(t *testing.T) {
	ts := newTestSet(t)
	g := GroupFromPublicKeys(ts.PubKeys)

	data := "somedata"

	sigs := make(SignatureMap)

	for i, signKey := range ts.SignKeys {
		sig, err := BlsSign(data, signKey)
		assert.Nil(t, err)
		sigs[ts.PubKeys[i].Id] = *sig
	}

	sig, err := g.CombineSignatures(sigs)
	assert.Nil(t, err)

	isVerified, err := g.VerifySignature(MustObjToHash(data), sig)
	assert.Nil(t, err)

	assert.True(t, isVerified)
}

func TestGroup_VerifySignature(t *testing.T) {
	ts := newTestSet(t)
	g := GroupFromPublicKeys(ts.PubKeys)
	data := "somedata"

	for _, test := range []struct {
		description  string
		generator    func(t *testing.T) (sigs SignatureMap)
		shouldVerify bool
	}{
		{
			description: "a valid signature",
			generator: func(t *testing.T) (sigs SignatureMap) {
				sigs = make(SignatureMap)

				for i, signKey := range ts.SignKeys {
					sig, err := BlsSign(data, signKey)
					assert.Nil(t, err)
					sigs[ts.PubKeys[i].Id] = *sig
				}
				return sigs
			},
			shouldVerify: true,
		},
		{
			description: "with only one signer",
			generator: func(t *testing.T) (sigs SignatureMap) {
				sigs = make(SignatureMap)
				i := 0
				sig, err := BlsSign(data, ts.SignKeys[i])
				assert.Nil(t, err)
				sigs[ts.PubKeys[i].Id] = *sig
				return sigs
			},
			shouldVerify: false,
		},
	} {
		sigs := test.generator(t)
		sig, err := g.CombineSignatures(sigs)
		assert.Nil(t, err)
		isVerified, err := g.VerifySignature(MustObjToHash(data), sig)

		assert.Equal(t, test.shouldVerify, isVerified, test.description)
	}
}