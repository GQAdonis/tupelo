package network

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	crypto "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-crypto"
	net "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-net"
	"github.com/quorumcontrol/qc3/p2p"
)

const ProtocolID = "tupelo/0.1"

var defaultBootstrapNodes = []string{}

type MessageParams struct {
	Source      *ecdsa.PrivateKey
	Destination *ecdsa.PublicKey
	Payload     []byte
}

type ReceivedMessage struct {
	Payload []byte

	Source      *ecdsa.PublicKey // Message recipient (identity used to decode the message)
	Destination *ecdsa.PublicKey // Message recipient (identity used to decode the message)
}

type Node struct {
	BoostrapNodes []string
	MessageChan   chan ReceivedMessage
	host          *p2p.Host
	key           *ecdsa.PrivateKey
	started       bool
	cancel        context.CancelFunc
}

// func (s *Subscription) RetrieveMessages() []*ReceivedMessage {
// 	whisperMsgs := s.whisperSubscription.Retrieve()
// 	msgs := make([]*ReceivedMessage, len(whisperMsgs))
// 	for i, msg := range whisperMsgs {
// 		msgs[i] = fromWhisper(msg)
// 	}
// 	return msgs
// }

func NewNode(key *ecdsa.PrivateKey) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	host, err := p2p.NewHost(ctx, key, 0)

	if err != nil {
		panic(fmt.Sprintf("Could not create node %v", err))
	}

	node := &Node{
		host:          host,
		key:           key,
		cancel:        cancel,
		BoostrapNodes: BootstrapNodes(),
		MessageChan:   make(chan ReceivedMessage, 10),
	}

	host.SetStreamHandler(ProtocolID, node.handler)

	return node
}

func BootstrapNodes() []string {
	if envSpecifiedNodes, ok := os.LookupEnv("TUPELO_BOOTSTRAP_NODES"); ok {
		return strings.Split(envSpecifiedNodes, ",")
	}

	return defaultBootstrapNodes
}

func (n *Node) Start() error {
	if n.started == true {
		return nil
	}

	_, err := n.host.Bootstrap(n.BoostrapNodes)
	if err != nil {
		return err
	}

	n.started = true
	return nil
}

func (n *Node) Stop() {
	if !n.started {
		return
	}
	n.cancel()
}

func (n *Node) Send(params MessageParams) error {
	fmt.Printf("sending %s\n", params.Payload)
	return n.host.Send(params.Destination, ProtocolID, params.Payload)
}

func (n *Node) handler(s net.Stream) {
	fmt.Printf("new stream from %v\n", s.Conn().RemotePeer().Pretty())
	data, err := ioutil.ReadAll(s)
	if err != nil {
		fmt.Printf("error reading: %v", err)
	}
	s.Close()
	fmt.Printf("received: %s\n", data)
	n.MessageChan <- ReceivedMessage{
		Payload: data,
		Source:  (*btcec.PublicKey)(s.Conn().RemotePublicKey().(*crypto.Secp256k1PublicKey)).ToECDSA(),
	}
}

func (n *Node) PublicKey() *ecdsa.PublicKey {
	return &n.key.PublicKey
}

// func (n *Node) SubscribeToTopic(topic []byte, symKey []byte) *Subscription {
// topicBytes := whisper.BytesToTopic(topic)

// sub := &Subscription{
// 	whisperSubscription: &whisper.Filter{
// 		KeySym:   symKey,
// 		Topics:   [][]byte{topicBytes[:]},
// 		AllowP2P: false,
// 	},
// }

// _, err := n.whisper.Subscribe(sub.whisperSubscription)
// if err != nil {
// 	panic(fmt.Sprintf("error subscribing: %v", err))
// }
// return sub
// }

// func (n *Node) SubscribeToKey(key *ecdsa.PrivateKey) *Subscription {
// sub := &Subscription{
// 	whisperSubscription: &whisper.Filter{
// 		AllowP2P: true,
// 		KeyAsym:  key,
// 	},
// }

// _, err := n.whisper.Subscribe(sub.whisperSubscription)
// if err != nil {
// 	panic(fmt.Sprintf("error subscribing: %v", err))
// }
// return sub
// }
