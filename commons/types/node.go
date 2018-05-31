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

// Set
func (this *Node) Set(txn *badger.Txn) error {
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
func (this *Node) Delete(txn *badger.Txn) error {
	err := txn.Delete([]byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

//func (this *Node) UnmarshalJSON(bytes []byte) error {
//	var jsonMap map[string]interface{}
//	err := json.Unmarshal(bytes, &jsonMap)
//	if err != nil {
//		return err
//	}
//	if jsonMap["address"] != nil {
//		this.Address = jsonMap["address"].(string)
//	}
//	if jsonMap["transaction"] != nil {
//		this.Transaction = jsonMap["transaction"].(Transaction)
//	}
//	if jsonMap["rumor"] != nil {
//		this.Rumors = jsonMap["rumor"].([]Rumor)
//	}
//	return nil
//}
//
//// MarshalJSON
//func (this Node) MarshalJSON() ([]byte, error) {
//	return json.Marshal(struct {
//		ReceiptId    string    		`json:"receiptId"`
//		Transaction  Transaction    `json:"transaction"`
//		Rumors       []Rumor    	`json:"rumor"`
//	}{
//		ReceiptId:    	this.ReceiptId,
//		Transaction:	this.Transaction,
//		Rumors:       	this.Rumors,
//	})
//}

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