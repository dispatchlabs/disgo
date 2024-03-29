package tree

import (
	"encoding/hex"
	"time"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	merkleTree "github.com/keybase/go-merkle-tree"
)

// TestObjFactory generates a bunch of test objects for debugging
type TestFactory struct {
	objs  map[string]merkleTree.KeyValuePair
	seqno int
}

type testValue struct {
	_struct bool `codec:",toarray"`
	Seqno   int
	Key     string
	KeyRaw  []byte
}

// NewTestObjFactor makes a new object factory for testing
func NewTestFactory() *TestFactory {
	return &TestFactory{
		objs: make(map[string]merkleTree.KeyValuePair),
	}
}

func (of TestFactory) dumpAll() []merkleTree.KeyValuePair {
	var ret []merkleTree.KeyValuePair
	for _, v := range of.objs {
		ret = append(ret, v)
	}
	return ret
}

// Mproduce makes many test objects.
func (of *TestFactory) Mproduce(n int) []merkleTree.KeyValuePair {
	var ret []merkleTree.KeyValuePair
	for i := 0; i < n; i++ {
		ret = append(ret, of.Produce())
	}
	return ret
}

// Produce one test object
func (of *TestFactory) Produce() merkleTree.KeyValuePair {
	var testTX *types.Transaction
	testTX = mockTransaction()
	key := crypto.GetHashBytes(testTX.Hash).Bytes()
	keyString := hex.EncodeToString(key)

	val := testValue{Seqno: of.seqno, Key: keyString, KeyRaw: key}
	of.seqno++
	kvp := merkleTree.KeyValuePair{Key: key, Value: val}
	of.objs[keyString] = kvp
	return kvp
}

func mockTransaction() *types.Transaction {
	key, _ := crypto.NewKey()

	tx, err := types.NewTransferTokensTransaction(
		key.GetPrivateKeyString(),
		key.Address,
		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
		1,
		0,
		time.Now().UnixNano(),
	)

	if err != nil {
		utils.Error("Could not create transaction %s", err.Error())
	}
	return tx
}

func (of *TestFactory) Construct() interface{} {
	return types.Transaction{}
}
