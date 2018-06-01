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
)

// Gossip
type Gossip struct {
	ReceiptId   string
	Transaction Transaction
	Rumors      []Rumor
}

// NewGossip -
func NewGossip(transaction Transaction, receipt Receipt) *Gossip {
	gossip := &Gossip{}
	gossip.ReceiptId = receipt.Id
	gossip.Transaction = transaction
	gossip.Rumors = []Rumor{}
	return gossip
}

// ToGossipFromJson -
func ToGossipFromJson(payload []byte) (*Gossip, error) {
	gossip := &Gossip{}
	err := json.Unmarshal(payload, gossip)
	if err != nil {
		return nil, err
	}
	return gossip, nil
}

// ToJsonByGossip
func ToJsonByGossip(gossip Gossip) ([]byte, error) {
	bytes, err := json.Marshal(gossip)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ToGossipByKey
func ToGossipByKey(txn *badger.Txn, key []byte) (*Gossip, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return nil, err
	}
	return gossip, err
}

// ToGossipByTransactionHash
func ToGossipByTransactionHash(txn *badger.Txn, transactionHash string) (*Gossip, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("table-gossip-%s", transactionHash)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return nil, err
	}
	return gossip, err
}

// ToGossips
func ToGossips(txn *badger.Txn) ([]*Gossip, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-gossip-"))
	var gossips = make([]*Gossip, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		gossip, err := ToGossipFromJson(value)
		if err != nil {
			utils.Error(err)
			continue
		}
		gossips = append(gossips, gossip)
	}
	return gossips, nil
}

func ToOldGossips(txn *badger.Txn) ([]*Gossip, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-post-gossip-"))
	var gossips = make([]*Gossip, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		gossip, err := ToGossipFromJson(value)
		if err != nil {
			utils.Error(err)
			continue
		}
		gossips = append(gossips, gossip)
	}
	return gossips, nil
}

// ContainsRumor
func (this Gossip) ContainsRumor(address string) bool {
	for _, persistedRumor := range this.Rumors {
		if persistedRumor.Address == address {
			return true
		}
	}
	return false
}

// ValidRumors
func (this Gossip) ValidRumors() int {
	validRumors := 0
	for _, rumor := range this.Rumors {
		if !rumor.Verify() {
			continue
		}
		validRumors++
	}
	return validRumors
}

// Key
func (this Gossip) Key() string {
	return fmt.Sprintf("table-gossip-%s", this.Transaction.Hash)
}

// Set
func (this *Gossip) Set(txn *badger.Txn) error {
	return txn.SetWithTTL([]byte(this.Key()), []byte(this.String()), GossipTTL)
}

// Refresh
func (this *Gossip) Refresh(txn *badger.Txn) error {
	item, err := txn.Get([]byte(this.Key()))
	if err != nil {
		return err
	}
	value, err := item.Value()
	if err != nil {
		return err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return err
	}
	this = gossip
	return nil
}

// String
func (this Gossip) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal gossip", err)
		return ""
	}
	return string(bytes)
}