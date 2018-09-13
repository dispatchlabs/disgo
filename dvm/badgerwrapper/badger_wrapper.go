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
package badgerwrapper

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/dvm/ethereum/common"

	// "github.com/dispatchlabs/disgo/dvm/ethereum/ethdb"
	"strings"

	"github.com/dispatchlabs/disgo/commons/crypto"
	disgoServices "github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/utils"
	ethdbInterfaces "github.com/dispatchlabs/disgo/dvm/ethereum/ethdb"
)

var badgerDatabaseInstance *BadgerDatabase
var badgerDatabaseOnce sync.Once

type BadgerDatabase struct {
}

func GetBadgerDatabase() *BadgerDatabase {
	badgerDatabaseOnce.Do(func() {
		badgerDatabaseInstance = &BadgerDatabase{}
	})

	return badgerDatabaseInstance
}

func NewBadgerDatabase() (*BadgerDatabase, error) {
	disgoServices.GetDbService()
	return GetBadgerDatabase(), nil
}

// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
// Database interface
// Based on https://github.com/dgraph-io/badger#using-keyvalue-pairs
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
func (db *BadgerDatabase) Put(key []byte, value []byte) error {
	utils.Debug(fmt.Sprintf("BadgerDatabase-PUT-Key   : %s", crypto.Encode(key)))
	utils.Debug(fmt.Sprintf("BadgerDatabase-PUT-KeyRAW: %v", key))
	// utils.Debug(fmt.Sprintf("BadgerDatabase-PUT-Val: %s", crypto.Encode(value)))

	// valEncoded := crypto.Encode(value)
	// if strings.Index(valEncoded, "f90163a0") > -1 {
	// 	utils.Debug("HERE!!!")
	// }

	err := disgoServices.GetDb().Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})

	return err
}

func (db *BadgerDatabase) Get(key []byte) ([]byte, error) {
	utils.Debug(fmt.Sprintf("BadgerDatabase-GET-Key   : %s", crypto.Encode(key)))
	utils.Debug(fmt.Sprintf("BadgerDatabase-GET-KeyRAW: %v", key))

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

		return nil
	})

	// utils.Debug(fmt.Sprintf("BadgerDatabase-GET-Val: %s", crypto.Encode(value)))
	return value, err
}

func (db *BadgerDatabase) Has(key []byte) (bool, error) {
	utils.Debug(fmt.Sprintf("BadgerDatabase-HAS-Key: %s", crypto.Encode(key)))

	item, err := db.Get(key)

	if err != nil {
		return false, err
	}

	return (item != nil), nil
}

func (db *BadgerDatabase) Delete(key []byte) error {
	utils.Debug(fmt.Sprintf("BadgerDatabase-DELETE-Key: %s", crypto.Encode(key)))

	return nil
}

func (db *BadgerDatabase) Close() {
	utils.Debug(fmt.Sprintf("BadgerDatabase-Close:"))
}

func (db *BadgerDatabase) NewBatch() ethdbInterfaces.Batch {
	utils.Debug(fmt.Sprintf("BadgerDatabase-NewBatch:"))
	return &memBatch{db: db}
}

func (db *BadgerDatabase) Dump() {
	//var items = make([]*proto.Item, 0)
	err := disgoServices.GetDb().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		fmt.Println("\n\n*****************************************************************************")
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			value, err := item.Value()
			if err != nil {
				return err
			}
			if !strings.HasPrefix(string(key), "key") {
				fmt.Printf("\nItem:\nKey: %v\nValue:%v\n", crypto.Encode(key), crypto.Encode(value))
				fmt.Printf("Key: %v\nValue: %v\n", string(key), string(value))
				fmt.Printf("Key: %v\nValue: %v\n", key, value)
			}
			//items = append(items, &proto.Item{Key: string(key), Value: string(value)})
		}
		fmt.Println("*****************************************************************************")
		return nil
	})
	if err != nil {
		utils.Error(err)
	}
	//for _, itm := range items {
	//	fmt.Printf("Item:\nKey: %v\nValue:%v\n", crypto.Encode([]byte(itm.Key)), itm.Value)
	//}
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
	b.writes = append(b.writes, kv{
		k: common.CopyBytes(key),
		v: common.CopyBytes(value),
	})

	b.size += len(value)
	return nil
}

func (b *memBatch) Delete(key []byte) error {
	newWrites := []kv{}
	for _, keyValues := range b.writes {
		if string(keyValues.k) != string(key) {
			newWrites = append(newWrites, kv{k: common.CopyBytes(keyValues.k), v: common.CopyBytes(keyValues.v)})
		} else {
			b.size -= len(keyValues.v)
		}
	}

	b.writes = newWrites

	return nil
}

func (b *memBatch) ValueSize() int {
	return b.size
}

func (b *memBatch) Write() error {
	for _, kv := range b.writes {

		// utils.Debug(fmt.Sprintf("memBatch-Write-KEY: %v", crypto.Encode(kv.k)))
		// utils.Debug(fmt.Sprintf("memBatch-Write-VAL: %v", crypto.Encode(kv.v)))

		// utils.Debug(fmt.Sprintf("memBatch-Write-KEY-RAW: %v", kv.k))
		// utils.Debug(fmt.Sprintf("memBatch-Write-VAL-RAW: %v", kv.v))

		b.db.Put(kv.k, kv.v)
	}

	return nil
}

func (b *memBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
