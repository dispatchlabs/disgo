package tree

import (

	"golang.org/x/net/context"

	"github.com/keybase/go-merkle-tree"
	"fmt"
	"github.com/dispatchlabs/disgo/commons/crypto"

	"github.com/gin-gonic/gin/json"
)

func Build() {

	// factory is an "object factory" that makes a whole bunch
	// of phony objects. Importantly, it fits the 'ValueConstructor'
	// interface, so that it can tell the MerkleTree class how
	// to pull type values out of the tree.
	factory := NewTestFactory()

	// Make a whole bunch of phony objects in our Object Factory.
	var objs []merkleTree.KeyValuePair
	objs = factory.Mproduce(48)

	// Collect and sort the objects into a "sorted map"
	var sm *merkleTree.SortedMap
	sm = merkleTree.NewSortedMapFromList(objs)

	// Make a test storage engine
	var eng merkleTree.StorageEngine
	eng = merkleTree.NewMemEngine()

	// 256 children per node; once there are 512 entries in a leaf,
	// then split the leaf by adding more parents.
	var config merkleTree.Config
	config = merkleTree.NewConfig(MyHasher{}, 16, 32, factory)

	// Make a new tree object with this engine and these config
	// values
	var mTree *merkleTree.Tree
	mTree = merkleTree.NewTree(eng, config)

	// Make an empty Tranaction info for now
	var txInfo merkleTree.TxInfo

	// Build the tree
	mTree.Build(context.TODO(), sm, txInfo)
	PrintMyTree(eng, mTree)
	//factory.ModifySome(20)
}

// Hasher is a simple hash function application
type MyHasher struct{}

// Hash the data
func (s MyHasher) Hash(b []byte) merkleTree.Hash {
	return crypto.NewHash(b).Bytes()
}

func PrintMyTree(eng merkleTree.StorageEngine, tree *merkleTree.Tree) {
	rootHash := tree.GetRoot(context.TODO())
	jsn, _ := rootHash.MarshalJSON()
	fmt.Printf("RootHash: = %v\n", string(jsn))

	rootNode, err := eng.LookupNode(context.TODO(), rootHash)
	if err != nil {
		panic(err)
	}
	var node *merkleTree.Node
	//var nodeExported []byte
	err = merkleTree.DecodeFromBytes(&node, rootNode)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	printNodes(tree, node)
}

func printNodes(tree *merkleTree.Tree, node *merkleTree.Node) {
	if(len(node.INodes) > 0) {
		for _, hash := range node.INodes {
			//var child *merkleTree.Node
			val, nodeHash, err := tree.Find(context.TODO(), hash)
			fmt.Printf("Node Value: %v\n", val)
			//err = merkleTree.DecodeFromBytes(&child, val)
			if err != nil {
				fmt.Errorf(err.Error())
			}
			jsn, _ := hash.MarshalJSON()
			jsn2, _ := nodeHash.MarshalJSON()
			fmt.Printf("Hash: = %v --> Parent %v\n", string(jsn), string(jsn2))
			//printNodes(tree, child)
		}
	}
	ret, _ := json.MarshalIndent(node, "", "   ")
	fmt.Printf("%v\n", string(ret))

}