package gossip2

import (
	"bytes"
	"fmt"
)

type conflictSetWorker struct {
	gn *GossipNode
}

func (csw *conflictSetWorker) HandleRequest(msg ProvideMessage, respCh chan processorResponse) {
	gn := csw.gn
	messageType := MessageType(msg.Key[8])
	conflictSetID := conflictSetIDFromMessageKey(msg.Key)

	log.Debugf("%s conflict set worker %s", gn.ID(), msg.Key)
	switch messageType {
	case MessageTypeSignature:
		log.Debugf("%v: handling a new Signature message", gn.ID())
		_, err := csw.HandleNewSignature(msg)
		if err != nil {
			respCh <- processorResponse{
				ConflictSetID: conflictSetID,
				Error:         err,
			}
		}
		respCh <- processorResponse{
			ConflictSetID: conflictSetID,
			NewSignature:  true,
		}
	case MessageTypeTransaction:
		log.Debugf("%v: handling a new Transaction message", gn.ID())
		didSign, err := csw.HandleNewTransaction(msg)
		if err != nil {
			respCh <- processorResponse{
				ConflictSetID: conflictSetID,
				Error:         err,
			}
		}
		respCh <- processorResponse{
			ConflictSetID:  conflictSetID,
			NewTransaction: didSign,
		}
	case MessageTypeDone:
		log.Debugf("%v: handling a new Done message", gn.ID())
		err := csw.handleDone(msg)
		if err != nil {
			respCh <- processorResponse{
				ConflictSetID: conflictSetID,
				Error:         err,
			}
		}
		respCh <- processorResponse{
			ConflictSetID: conflictSetID,
			IsDone:        true,
		}
	default:
		log.Errorf("%v: unknown message %v", gn.ID(), msg.Key)
	}
}

func (csw *conflictSetWorker) handleDone(msg ProvideMessage) error {
	gn := csw.gn
	log.Debugf("%s handling done message %v", gn.ID(), msg.Key)

	conflictSetID := conflictSetIDFromMessageKey(msg.Key)

	conflictSetKeys, err := gn.Storage.GetKeysByPrefix(conflictSetID[0:4])

	for _, key := range conflictSetKeys {
		if key[8] != byte(MessageTypeDone) && bytes.Equal(conflictSetIDFromMessageKey(key), conflictSetID) {
			gn.Remove(key)
			if err != nil {
				log.Errorf("%s error deleting conflict set item %v", gn.ID(), key)
			}
		}
	}

	return nil
}

func (csw *conflictSetWorker) HandleNewSignature(msg ProvideMessage) (accepted bool, err error) {
	log.Debugf("%s handling sig message %v", csw.gn.ID(), msg.Key)
	// TOOD: actually check to see if we accept this sig
	return true, nil
}

func (csw *conflictSetWorker) HandleNewTransaction(msg ProvideMessage) (didSign bool, err error) {
	gn := csw.gn
	//TODO: check if transaction hash matches key hash
	//TODO: check if the conflict set is done
	//TODO: sign this transaction if it's valid and new

	var t Transaction
	_, err = t.UnmarshalMsg(msg.Value)

	if err != nil {
		return false, fmt.Errorf("error getting transaction: %v", err)
	}
	log.Debugf("%s new transaction %s", gn.ID(), bytesToString(t.ID()))

	isValid, err := csw.IsTransactionValid(t)
	if err != nil {
		return false, fmt.Errorf("error validating transaction: %v", err)
	}

	if isValid {
		//TODO: sign the right thing
		sig, err := gn.SignKey.Sign(msg.Value)
		if err != nil {
			return false, fmt.Errorf("error signing key: %v", err)
		}
		signature := Signature{
			TransactionID: t.ID(),
			Signers:       map[string]bool{gn.address: true},
			Signature:     sig,
		}
		encodedSig, err := signature.MarshalMsg(nil)
		if err != nil {
			return false, fmt.Errorf("error marshaling sig: %v", err)
		}
		log.Debugf("%s signing %v", gn.ID(), signature.StoredID(t.ToConflictSet().ID()))
		sigID := signature.StoredID(t.ToConflictSet().ID())
		gn.Add(sigID, encodedSig)

		sigMessage := ProvideMessage{
			Key:   sigID,
			Value: encodedSig,
		}
		gn.newObjCh <- sigMessage
		return true, nil
	}

	gn.Remove(msg.Key)
	log.Errorf("%s error, invalid transaction", gn.ID())
	return false, nil
}

func (csw *conflictSetWorker) IsTransactionValid(t Transaction) (bool, error) {
	// state, err := gn.Storage.Get(t.ObjectID)
	// if err != nil {
	// 	return false, fmt.Errorf("error getting state: %v", err)
	// }

	// here we would send transaction and state to a handler
	return true, nil
}