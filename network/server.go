package network

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ipfs/go-ipld-cbor"
	"github.com/quorumcontrol/chaintree/dag"
	"sync"
	"time"
)

const DefaultTTL = 60

func init() {
	cbornode.RegisterCborType(Request{})
	cbornode.RegisterCborType(Response{})
}

type Request struct {
	Type    string
	Id      string
	Payload interface{}
	dst     *ecdsa.PublicKey
	src     *ecdsa.PublicKey
}

type Response struct {
	Id      string
	Code    int
	Payload interface{}
}

type HandlerFunc func(req Request) (*Response, error)

type RequestHandler struct {
	node        *Node
	mappings    map[string]HandlerFunc
	subs        []*Subscription
	lock        *sync.Mutex
	closeChan   chan bool
	messageChan chan *Request
}

func NewRequestHandler(client *Node) *RequestHandler {
	return &RequestHandler{
		node:        client,
		mappings:    make(map[string]HandlerFunc),
		closeChan:   make(chan bool, 2),
		messageChan: make(chan *Request, 5),
		lock:        &sync.Mutex{},
	}
}

func (rh *RequestHandler) AssignHandler(requestType string, handlerFunc HandlerFunc) error {
	rh.lock.Lock()
	defer rh.lock.Unlock()

	rh.mappings[requestType] = handlerFunc
	return nil
}

func (rh *RequestHandler) HandleTopic(topic []byte, symkey []byte) {
	rh.lock.Lock()
	defer rh.lock.Unlock()

	rh.subs = append(rh.subs, rh.node.SubscribeToTopic(topic, symkey))
}

func (rh *RequestHandler) HandleKey(key *ecdsa.PrivateKey) {
	rh.lock.Lock()
	defer rh.lock.Unlock()

	rh.subs = append(rh.subs, rh.node.SubscribeToKey(key))
}

func (rh *RequestHandler) Start() {

	go func() {
		for {
			select {
			case req := <-rh.messageChan:
				log.Debug("request received", "type", req.Type)
				handler, ok := rh.mappings[req.Type]
				if ok {
					resp, err := handler(*req)
					resp.Id = req.Id
					if err != nil {
						log.Error("error handling message", "err", err)
						break
					}
					sw := &dag.SafeWrap{}
					node := sw.WrapObject(resp)
					if sw.Err != nil {
						log.Error("error wrapping response", "err", sw.Err)
						break
					}

					rh.node.Send(MessageParams{
						Payload: node.RawData(),
						TTL:     DefaultTTL,
						PoW:     0.1,
						Dst:     req.src,
						Src:     rh.node.key,
					})
				} else {
					log.Info("invalid message type", "type", req.Type)
				}
			case <-rh.closeChan:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				for _, sub := range rh.subs {
					messages := sub.RetrieveMessages()
					for _, msg := range messages {
						log.Debug("message received", "msg", msg)
						rh.messageChan <- messageToRequest(msg)
					}
				}
			case <-rh.closeChan:
				return
			}
		}
	}()
}

func (rh *RequestHandler) Stop() {
	rh.closeChan <- true
	rh.closeChan <- true
}

func messageToRequest(message *ReceivedMessage) *Request {
	req := &Request{}
	err := cbornode.DecodeInto(message.Payload, req)
	if err != nil {
		log.Error("invalid message", "err", err)
		return nil
	}
	return req
}
