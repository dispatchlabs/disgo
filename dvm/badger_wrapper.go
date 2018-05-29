/*
 *    This file is part of DVM library.
 *
 *    The DVM library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DVM library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DVM library.  If not, see <http://www.gnu.org/licenses/>.
*/
package dvm

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/dvm/ethereum/common"
	// "github.com/dispatchlabs/dvm/ethereum/ethdb"
	disgoServices "github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/utils"
	ethdbInterfaces "github.com/dispatchlabs/dvm/ethereum/ethdb"
)

type BadgerDatabase struct {
}

func NewBadgerDatabase() (*BadgerDatabase, error) {
	disgoServices.GetDbService()
	return &BadgerDatabase{}, nil
}

// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
// Database interface
// Based on https://github.com/dgraph-io/badger#using-keyvalue-pairs
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
func (db *BadgerDatabase) Put(key []byte, value []byte) error {
	// utils.Info(fmt.Sprintf("K: %s, V: %s", string(key), string(value)))

	// var returnE error = nil

	// txn := disgoServices.NewTxn(true)
	// defer txn.Discard()

	// txn.Set(key, value)
	// txn.Commit(func(e error) {
	// 	returnE = e
	// })

	// return returnE

	err := disgoServices.GetDb().Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})

	return err
}

func (db *BadgerDatabase) Get(key []byte) ([]byte, error) {
	// utils.Info(fmt.Sprintf("K: %s", string(key)))

	// txn := disgoServices.NewTxn(false)
	// defer txn.Discard()

	// item, err := txn.Get(key)
	// if err != nil {
	// 	return nil, err
	// }

	// return item.Value()

	var value []byte
	err := disgoServices.GetDb().View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		value = make([]byte, len(val))
		copy(value, val[:])

		// fmt.Printf("The answer is: %s\n", val)
		return nil
	})

	return value, err
}

func (db *BadgerDatabase) Has(key []byte) (bool, error) {
	// utils.Info(fmt.Sprintf("K: %s", string(key)))

	item, err := db.Get(key)

	if err != nil {
		return false, err
	}

	return (item != nil), nil
}

func (db *BadgerDatabase) Delete(key []byte) error {
	utils.Info(fmt.Sprintf("K: %s", string(key)))

	return nil
}

func (db *BadgerDatabase) Close() {
	utils.Info("BadgerDatabase/Close")
}

func (db *BadgerDatabase) NewBatch() ethdbInterfaces.Batch {
	return &memBatch{db: db}
}

// func (db *BadgerDatabase) Keys() [][]byte {
// 	db.lock.RLock()
// 	defer db.lock.RUnlock()

// 	keys := [][]byte{}
// 	for key := range db.db {
// 		keys = append(keys, []byte(key))
// 	}
// 	return keys
// }

// func (db *BadgerDatabase) Len() int { return len(db.db) }

// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
// Batch interface
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~

type kv struct{ k, v []byte }

type memBatch struct {
	db     *BadgerDatabase
	writes []kv
	size   int
}

func (b *memBatch) Put(key, value []byte) error {
	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value)})
	b.size += len(value)
	return nil
}

func (b *memBatch) ValueSize() int {
	return b.size
}

func (b *memBatch) Write() error {
	for _, kv := range b.writes {
		b.db.Put(kv.k, kv.v)
	}

	return nil
}

func (b *memBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
