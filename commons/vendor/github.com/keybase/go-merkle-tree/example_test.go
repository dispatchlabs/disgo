package merkleTree

import (
	"golang.org/x/net/context"
)

func ExampleTree_Build() {

	// factory is an "object factory" that makes a whole bunch
	// of phony objects. Importantly, it fits the 'ValueConstructor'
	// interface, so that it can tell the MerkleTree class how
	// to pull type values out of the tree.
	factory := NewTestObjFactory()

	// Make a whole bunch of phony objects in our Object Factory.
	var objs []KeyValuePair
	objs = factory.Mproduce(1024)

	// Collect and sort the objects into a "sorted map"
	var sm *SortedMap
	sm = NewSortedMapFromList(objs)

	// Make a test storage engine
	var eng StorageEngine
	eng = NewMemEngine()

	// 256 children per node; once there are 512 entries in a leaf,
	// then split the leaf by adding more parents.
	var config Config
	config = NewConfig(SHA512Hasher{}, 256, 512, factory)

	// Make a new tree object with this engine and these config
	// values
	var tree *Tree
	tree = NewTree(eng, config)

	// Make an empty Tranaction info for now
	var txInfo TxInfo

	// Build the tree
	tree.Build(context.TODO(), sm, txInfo)
}
