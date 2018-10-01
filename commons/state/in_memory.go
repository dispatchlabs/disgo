package state

// import (
// 	"github.com/dispatchlabs/disgo/commons/tree"
// 	"github.com/dispatchlabs/disgo/commons/utils"
// )

// func NewMerkleTree(content []tree.MerkleTreeContent) *tree.MerkleTree {

// 	merkleTree, err := tree.NewTree(content)
// 	if err != nil {
// 		utils.Error(err.Error())
// 	}

// 	utils.Info(merkleTree.String())
// 	return merkleTree
// }

/*
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Commit(onleaf trie.LeafCallback) (crypto.HashBytes, error)
	Hash() crypto.HashBytes
	NodeIterator(startKey []byte) trie.NodeIterator
	GetKey([]byte) []byte // TODO(fjl): remove this when SecureTrie is removed
	Prove(key []byte, fromLevel uint, proofDb ethdb.Putter) error

*/
