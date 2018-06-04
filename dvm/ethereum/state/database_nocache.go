// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"fmt"
	"sync"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/ethdb"
	"github.com/dispatchlabs/disgo/dvm/ethereum/trie"
)

func NewNonCacheDatabase(db ethdb.Database) Database {
	return &nonCacheDB{
		db: trie.NewDatabase(db),
	}
}

type nonCacheDB struct {
	db *trie.Database
	mu sync.Mutex
}

// OpenTrie opens the main account trie.
func (db *nonCacheDB) OpenTrie(root crypto.HashBytes) (Trie, error) {
	utils.Info(fmt.Sprintf("nonCacheDB-OpenTrie: %s", crypto.Encode(root[:])))

	db.mu.Lock()
	defer db.mu.Unlock()

	tr, err := trie.NewSecure(root, db.db, MaxTrieCacheGen)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (db *nonCacheDB) pushTrie(t *trie.SecureTrie) {
	utils.Info(fmt.Sprintf("nonCacheDB-pushTrie: %v", t))

	db.mu.Lock()
	defer db.mu.Unlock()

	// if len(db.pastTries) >= maxPastTries {
	// 	copy(db.pastTries, db.pastTries[1:])
	// 	db.pastTries[len(db.pastTries)-1] = t
	// } else {
	// 	db.pastTries = append(db.pastTries, t)
	// }
}

// OpenStorageTrie opens the storage trie of an account.
func (db *nonCacheDB) OpenStorageTrie(addrHash, root crypto.HashBytes) (Trie, error) {
	utils.Info(fmt.Sprintf("nonCacheDB-OpenStorageTrie: %s", crypto.Encode(root[:])))

	return trie.NewSecure(root, db.db, 0)
}

// CopyTrie returns an independent copy of the given trie.
func (db *nonCacheDB) CopyTrie(t Trie) Trie {
	utils.Info(fmt.Sprintf("nonCacheDB-CopyTrie: %v", t))

	switch t := t.(type) {
	// case cachedTrie:
	// 	return cachedTrie{t.SecureTrie.Copy(), db}
	case *trie.SecureTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

// ContractCode retrieves a particular contract's code.
func (db *nonCacheDB) ContractCode(addrHash, codeHash crypto.HashBytes) ([]byte, error) {
	utils.Info(fmt.Sprintf("nonCacheDB-ContractCode: %s -> %s", crypto.Encode(addrHash[:]), crypto.Encode(codeHash[:])))

	code, err := db.db.Node(codeHash)
	return code, err
}

// ContractCodeSize retrieves a particular contracts code's size.
func (db *nonCacheDB) ContractCodeSize(addrHash, codeHash crypto.HashBytes) (int, error) {
	utils.Info(fmt.Sprintf("nonCacheDB-ContractCodeSize: %s -> %s", crypto.Encode(addrHash[:]), crypto.Encode(codeHash[:])))

	code, err := db.ContractCode(addrHash, codeHash)
	return len(code), err
}

// TrieDB retrieves any intermediate trie-node caching layer.
func (db *nonCacheDB) TrieDB() *trie.Database {
	utils.Info(fmt.Sprintf("nonCacheDB-TrieDB:"))

	return db.db
}
