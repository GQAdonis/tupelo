package gossip2

import (
	"bytes"

	"github.com/ethereum/go-ethereum/crypto"
)

var doneBytes = []byte("done")

func init() {
	if len(doneBytes) != 4 {
		panic("doneBytes must be 4 bytes!")
	}
}

func (c *ConflictSet) ID() []byte {
	return crypto.Keccak256(concatBytesSlice(c.ObjectID, c.Tip))
}

func doneIDFromConflictSetID(conflictSetID []byte) []byte {
	return concatBytesSlice(conflictSetID[0:4], doneBytes, []byte{byte(MessageTypeDone)}, conflictSetID)
}

func (c *ConflictSet) DoneID() []byte {
	return doneIDFromConflictSetID(c.ID())
}

func isDone(storage *BadgerStorage, conflictSetID []byte) (bool, error) {
	doneID := doneIDFromConflictSetID(conflictSetID)
	doneIDPrefix := doneID[0:8]

	conflictSetDoneKeys, err := storage.GetKeysByPrefix(doneIDPrefix)
	if err != nil {
		return false, err
	}
	for _, key := range conflictSetDoneKeys {
		if bytes.Equal(key, doneID) {
			return true, nil
		}
	}

	return false, nil
}