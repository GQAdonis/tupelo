package consensus

import (
	"crypto/ecdsa"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-cbor"
	"github.com/quorumcontrol/chaintree/chaintree"
)

func init() {
	cbornode.RegisterCborType(AddBlockRequest{})
	cbornode.RegisterCborType(AddBlockResponse{})
	cbornode.RegisterCborType(FeedbackRequest{})
	cbornode.RegisterCborType(TipRequest{})
	cbornode.RegisterCborType(TipResponse{})
}

type Wallet interface {
	GetChain(id string) (*SignedChainTree, error)
	SaveChain(chain *SignedChainTree) error
	//Balances(id string) (map[string]uint64, error)
	GetChainIds() ([]string, error)
	Close()

	GetKey(addr string) (*ecdsa.PrivateKey, error)
	GenerateKey() (*ecdsa.PrivateKey, error)
	ListKeys() ([]string, error)
}

type AddBlockRequest struct {
	Nodes    [][]byte
	Tip      *cid.Cid
	NewBlock *chaintree.BlockWithHeaders
}

type AddBlockResponse struct {
	SignerId  string
	ChainId   string
	Tip       *cid.Cid
	Signature Signature
}

type FeedbackRequest struct {
	ChainId   string
	Tip       *cid.Cid
	Signature Signature
}

type TipRequest struct {
	ChainId string
}

type TipResponse struct {
	ChainId   string
	Tip       *cid.Cid
	Signature Signature
}