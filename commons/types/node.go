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
	//"reflect"
	"strings"
	"encoding/hex"
	"bytes"
	"encoding/binary"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"errors"
)

// Node - Is the DisGover's notion of what a node is
type Node struct {
	Hash         string    `json:"hash"`
	Address      string    `json:"address"`
	Signature    string    `json:"signature"`
	GrpcEndpoint *Endpoint `json:"grpcEndpoint"`
	HttpEndpoint *Endpoint `json:"httpEndpoint"`
	Type         string    `json:"type"`
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
func (this *Node) Cache(cache *cache.Cache, time_optional ...time.Duration) {
	TTL := NodeTTL
	if len(time_optional) > 0 {
		TTL = time_optional[0]
	}
	cache.Set(this.Key(), this, TTL)
	cache.Set(this.TypeKey(), this.Key(), TTL)
}

//Persist
func (this *Node) Persist(txn *badger.Txn) error {
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

// PersistAndCache
func (this *Node) PersistAndCache(txn *badger.Txn, cache *cache.Cache, ttl time.Duration) error {
	this.Cache(cache, ttl)
	return this.Persist(txn)
}

// Unset
func (this *Node) Unset(txn *badger.Txn, cache *cache.Cache) error {
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

// NewHash
func (this Node) NewHash() (string, error) {
	addressBytes, err := hex.DecodeString(this.Address)
	if err != nil {
		utils.Error("unable decode address", err)
		return "", err
	}
	var values = []interface{}{
		addressBytes,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Error("unable to write node bytes to buffer", err)
			return "", err
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	return hex.EncodeToString(hash[:]), nil
}

// NewSignature
func (this Node) NewSignature(privateKey string) (string, error) {
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return "", err
	}
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		utils.Error("unable to decode privateKey", err)
		return "", err
	}
	signatureBytes, err := crypto.NewSignature(privateKeyBytes, hashBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signatureBytes), nil
}

// Verify
func (this Node) Verify() error {

	// Hash ok?
	hash, err := this.NewHash()
	if err != nil {
		return errors.New("unable to compute hash")
	}
	if this.Hash != hash {
		return errors.New("invalid hash")
	}

	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return errors.New("unable to decode hash")
	}
	signatureBytes, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Error("unable to decode signature", err)
		return errors.New("unable to decode signature")
	}
	publicKeyBytes, err := crypto.ToPublicKey(hashBytes, signatureBytes)
	if err != nil {
		utils.Error("unable to generate public key from hash and signature", err)
		return errors.New("unable to generate public key from hash and signature")
	}

	// Derived address from publicKeyBytes match from?
	address := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if address != this.Address {
		return errors.New("node address does not match the computed address from hash and signature")
	}
	if !crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes) {
		return errors.New("invalid signature")
	}

	return nil
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
	value, ok := cache.Get(fmt.Sprintf("table-node-%s", address))
	if !ok {
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

// ToNodeByTypeFromCache -
func ToNodesByTypeFromCache(c *cache.Cache, tipe string) ([]*Node, error) {
	var keys []string
	var nodes []*Node
	for i, value := range c.Items() {
		useful := strings.HasPrefix(i, fmt.Sprintf("key-node-type-%s", tipe))
		if useful {
			key := value.Object.(string)
			keys = append(keys, key)
		}
	}
	for i, value := range c.Items() {
		for _, key := range keys {
			if i == key {
				node := value.Object.(*Node)
				nodes = append(nodes, node)
			}
		}
	}
	return nodes, nil
}
