/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package types

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
	"time"
)

// Node - Is the DisGover's notion of what a node is
type Node struct {
	Address  string    `json:"address"`
	Endpoint *Endpoint `json:"endpoint"`
	Type     string    `json:"type"`
}

// Key
func (this Node) Key() string {
	return fmt.Sprintf("table-node-%s", this.Address)
}

// TypeKey
func (this Node) TypeKey() string {
	return fmt.Sprintf("key-node-type-%s-%s", this.Type, this.Address)
}

//Cache
func (this *Node) Cache(cache *cache.Cache, time_optional ...time.Duration){
	TTL := NodeTTL
	if len(time_optional) > 0 {
		TTL = time_optional[0]
	}
	cache.Set(this.Key(), this, TTL)
}

//Persist
func (this *Node) Persist(txn *badger.Txn) error{
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.TypeKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// Set
func (this *Node) Set(txn *badger.Txn,cache *cache.Cache) error {
	this.Cache(cache)
	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

// Unset
func (this *Node) Unset(txn *badger.Txn,cache *cache.Cache) error {
	cache.Delete(this.Key())
	err := txn.Delete([]byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}


// String
func (this Node) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal node", err)
		return ""
	}
	return string(bytes)
}

// ToTransactionFromJson -
func ToNodeFromJson(payload []byte) (*Node, error) {
	node := &Node{}
	err := json.Unmarshal(payload, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ToGossipFromCache -
func ToNodeFromCache(cache *cache.Cache, address string) (*Node, error) {
	value, ok :=cache.Get(fmt.Sprintf("table-node-%s", address))
	if !ok{
		return nil, ErrNotFound
	}
	node := value.(*Node)
	return node, nil
}

// ToNodeByKey
func ToNodeByKey(txn *badger.Txn, key []byte) (*Node, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	node, err := ToNodeFromJson(value)
	if err != nil {
		return nil, err
	}
	return node, err
}

// ToNodeByAddress
func ToNodeByAddress(txn *badger.Txn, address string) (*Node, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("table-node-%s", address)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	node, err := ToNodeFromJson(value)
	if err != nil {
		return nil, err
	}
	return node, err
}

// ToNodesByType
func ToNodesByType(txn *badger.Txn, tipe string) ([]*Node, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-node-type-%s", tipe))
	var nodes = make([]*Node, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		node, err := ToNodeByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}