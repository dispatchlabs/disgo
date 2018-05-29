package tree

import (
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/commons/crypto"
	"github.com/dispatchlabs/commons/types"
	merkleTree "github.com/keybase/go-merkle-tree"
	"crypto/rand"
	"time"
	"encoding/hex"

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
	key := testTX.CalculateHash()
	keyString := hex.EncodeToString(key)

	val := testValue{Seqno: of.seqno, Key: keyString, KeyRaw: key}
	of.seqno++
	kvp := merkleTree.KeyValuePair{Key: key, Value: val}
	of.objs[keyString] = kvp
	return kvp
}

func mockTransaction() *types.Transaction {
	key, _ := crypto.NewKey(rand.Reader)

	tx, err := types.NewTransaction(
		key.GetPrivateKeyString(),
		0,
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
