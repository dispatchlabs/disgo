package tree

// import (
// 	"golang.org/x/net/context"

// 	merkleTree "github.com/keybase/go-merkle-tree"
// 	"fmt"
// )

// func ExampleTree_Build() {

// 	// factory is an "object factory" that makes a whole bunch
// 	// of phony objects. Importantly, it fits the 'ValueConstructor'
// 	// interface, so that it can tell the MerkleTree class how
// 	// to pull type values out of the tree.
// 	factory := merkleTree.NewTestObjFactory()

// 	// Make a whole bunch of phony objects in our Object Factory.
// 	var objs []merkleTree.KeyValuePair
// 	objs = factory.Mproduce(1024)

// 	// Collect and sort the objects into a "sorted map"
// 	var sm *merkleTree.SortedMap
// 	sm = merkleTree.NewSortedMapFromList(objs)

// 	// Make a test storage engine
// 	var eng merkleTree.StorageEngine
// 	eng = merkleTree.NewMemEngine()

// 	// 256 children per node; once there are 512 entries in a leaf,
// 	// then split the leaf by adding more parents.
// 	var config merkleTree.Config
// 	config = merkleTree.NewConfig(merkleTree.SHA512Hasher{}, 256, 512, factory)

// 	// Make a new tree object with this engine and these config
// 	// values
// 	var tree *merkleTree.Tree
// 	tree = merkleTree.NewTree(eng, config)

// 	// Make an empty Tranaction info for now
// 	var txInfo merkleTree.TxInfo

// 	// Build the tree
// 	tree.Build(context.TODO(), sm, txInfo)
// 	PrintTree(eng, tree)
// 	factory.ModifySome(20)
// }

// func PrintTree(eng merkleTree.StorageEngine, tree *merkleTree.Tree) {
// 	rootHash := tree.GetRoot(context.TODO())
// 	jsn, _ := rootHash.MarshalJSON()
// 	fmt.Printf("RootHash: = %v\n", string(jsn))

// 	rootNode, err := eng.LookupNode(context.TODO(), rootHash)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Printf("Root: = %v\n", string(rootNode))
// }