package gossipclient

import (
	"crypto/ecdsa"

	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ipfs/go-ipld-cbor"
	"github.com/quorumcontrol/chaintree/chaintree"
	"github.com/quorumcontrol/qc3/consensus"
	"github.com/quorumcontrol/qc3/network"
)

type Handler interface {
	DoRequest(dst *ecdsa.PublicKey, req *network.Request) (chan *network.Response, error)
	AssignHandler(requestType string, handlerFunc network.HandlerFunc) error
	Start()
	Stop()
}

type GossipClient struct {
	handler    Handler
	started    bool
	SessionKey *ecdsa.PrivateKey
	Group      *consensus.Group
}

func NewGossipClient(group *consensus.Group) *GossipClient {
	sessionKey, err := crypto.GenerateKey()
	if err != nil {
		panic("error generating key")
	}

	node := network.NewNode(sessionKey)

	return &GossipClient{
		SessionKey: sessionKey,
		handler:    network.NewMessageHandler(node, []byte(group.Id())),
		Group:      group,
	}
}

func (gc *GossipClient) Start() {
	if gc.started {
		return
	}
	gc.started = true

	gc.handler.Start()
}

func (gc *GossipClient) Stop() {
	if !gc.started {
		return
	}
	gc.started = false
	gc.handler.Stop()
}

func (gc *GossipClient) TipRequest(chainId string) (*consensus.TipResponse, error) {
	rn := gc.Group.RandomMember()
	req, err := network.BuildRequest(consensus.MessageType_TipRequest, &consensus.TipRequest{
		ChainId: chainId,
	})
	if err != nil {
		return nil, fmt.Errorf("error building requeset: %v", err)
	}

	respChan, err := gc.handler.DoRequest(crypto.ToECDSAPub(rn.DstKey.PublicKey), req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %v", err)
	}

	respBytes := <-respChan
	if respBytes.Code == 200 {
		tipResp := &consensus.TipResponse{}
		err = cbornode.DecodeInto(respBytes.Payload, tipResp)
		if err != nil {
			return nil, fmt.Errorf("error decoding: %v", err)
		}
		return tipResp, nil
	} else {
		return nil, fmt.Errorf("error on request code: %d, err: %s", respBytes.Code, respBytes.Payload)
	}
}

func (gc *GossipClient) PlayTransactions(tree *consensus.SignedChainTree, treeKey *ecdsa.PrivateKey, remoteTip string, transactions []*chaintree.Transaction) (*consensus.AddBlockResponse, error) {

	unsignedBlock := &chaintree.BlockWithHeaders{
		Block: chaintree.Block{
			PreviousTip:  remoteTip,
			Transactions: transactions,
		},
	}

	blockWithHeaders, err := consensus.SignBlock(unsignedBlock, treeKey)
	if err != nil {
		return nil, fmt.Errorf("error signing: %v", err)
	}

	//TODO: only send the necessary nodes
	nodes := make([][]byte, len(tree.ChainTree.Dag.Nodes()))
	for i, node := range tree.ChainTree.Dag.Nodes() {
		nodes[i] = node.Node.RawData()
	}

	addBlockRequest := &consensus.AddBlockRequest{
		Nodes:    nodes,
		NewBlock: blockWithHeaders,
		Tip:      tree.Tip(),
		ChainId:  tree.MustId(),
	}

	log.Debug("sending: ", "tip", addBlockRequest.Tip, "nodeLength", len(nodes))

	req, err := network.BuildRequest(consensus.MessageType_AddBlock, addBlockRequest)

	rn := gc.Group.RandomMember()

	respChan, err := gc.handler.DoRequest(crypto.ToECDSAPub(rn.DstKey.PublicKey), req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %v", err)
	}

	isValid, err := tree.ChainTree.ProcessBlock(blockWithHeaders)
	if err != nil || !isValid {
		return nil, fmt.Errorf("error, invalid transactions: %v", err)
	}

	resp := <-respChan

	if resp.Code == 200 {
		addResponse := &consensus.AddBlockResponse{}
		err = cbornode.DecodeInto(resp.Payload, addResponse)
		if err != nil {
			return nil, fmt.Errorf("error decoding: %v", err)
		}
		return addResponse, nil
	} else {
		return nil, fmt.Errorf("error on request code: %d, err: %s", resp.Code, resp.Payload)
	}
}