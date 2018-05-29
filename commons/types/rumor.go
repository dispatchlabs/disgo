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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/commons/crypto"
	"github.com/dispatchlabs/commons/utils"
	"time"
)

// Rumor
type Rumor struct {
	Hash            string // Hash = (Address + TransactionHash + Time)
	Address         string
	TransactionHash string
	Time            int64
	Signature       string
}

// ToRumorFromJson -
func ToRumorFromJson(payload []byte) (*Rumor, error) {
	rumor := &Rumor{}
	err := json.Unmarshal(payload, rumor)
	if err != nil {
		return nil, err
	}
	return rumor, nil
}

// ToRumorsFromJson -
func ToRumorsFromJson(payload []byte) ([]*Rumor, error) {
	var rumors = make([]*Rumor, 0)
	err := json.Unmarshal(payload, &rumors)
	if err != nil {
		return nil, err
	}
	return rumors, nil
}

// NewRumor -
func NewRumor(privateKey string, address string, transactionHash string) *Rumor {
	rumor := &Rumor{}
	rumor.Address = address
	rumor.TransactionHash = transactionHash
	rumor.Time = utils.ToMilliSeconds(time.Now())
	rumor.Hash = rumor.NewHash()
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		utils.Error("unable to decode privateKey", err)
		return nil
	}
	hashBytes, err := hex.DecodeString(rumor.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return nil
	}
	signature, err := crypto.NewSignature(privateKeyBytes, hashBytes)
	if err != nil {
		utils.Error(err.Error())
		return nil
	}

	rumor.Signature = hex.EncodeToString(signature)
	return rumor
}

// NewHash
func (this Rumor) NewHash() string {
	addressBytes, err := hex.DecodeString(this.Address)
	if err != nil {
		utils.Error("unable to decode address", err)
		return ""
	}
	transactionHashBytes, err := hex.DecodeString(this.TransactionHash)
	if err != nil {
		utils.Error("unable to decode transaction", err)
		return ""
	}
	var values = []interface{}{
		addressBytes,
		transactionHashBytes,
		this.Time,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Fatal("unable to write rumor bytes to buffer", err)
			return ""
		}
	}
	delegateHash := crypto.NewHash(buffer.Bytes())
	return hex.EncodeToString(delegateHash[:])
}

// Verify
func (this Rumor) Verify() bool {
	if len(this.Hash) != crypto.HashLength*2 {
		return false
	}
	if len(this.Address) != crypto.AddressLength*2 {
		return false
	}
	if len(this.TransactionHash) != crypto.HashLength*2 {
		return false
	}
	if len(this.Signature) != crypto.SignatureLength*2 {
		return false
	}

	// Hash ok?
	if this.Hash != this.NewHash() {
		return false
	}
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return false
	}
	signatureBytes, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Error("unable to decode signature", err)
		return false
	}
	publicKeyBytes, err := crypto.ToPublicKey(hashBytes, signatureBytes)
	if err != nil {
		return false
	}

	// Derived address from publicKeyBytes match address?
	address := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if address != this.Address {
		return false
	}
	return crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes)
}

// ToRumorByKey
func ToRumorByKey(txn *badger.Txn, key []byte) (*Rumor, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	transaction, err := ToRumorFromJson(value)
	if err != nil {
		return nil, err
	}
	return transaction, err
}

// ToRumors
func ToRumors(txn *badger.Txn) ([]*Rumor, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-rumor-"))
	var transactions = make([]*Rumor, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToRumorFromJson(value)
		if err != nil {
			utils.Error(err)
			continue
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToRumorByAddress
func ToRumorsByAddress(txn *badger.Txn, address string) ([]*Rumor, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-rumor-address-%s", address))
	var rumors = make([]*Rumor, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		rumor, err := ToRumorByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		rumors = append(rumors, rumor)
	}
	return rumors, nil
}

// ToRumorsByTransactionHash
func ToRumorsByTransactionHash(txn *badger.Txn, transactionHash string) ([]*Rumor, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-rumor-hash-%s", transactionHash))
	var rumors = make([]*Rumor, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		rumor, err := ToRumorByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		rumors = append(rumors, rumor)
	}
	return rumors, nil
}

// ToJsonByRumorTransactionHash
func ToJsonByRumorTransactionHash(txn *badger.Txn, transactionHash string) ([]byte, error) {
	rumors, err := ToRumorsByTransactionHash(txn, transactionHash)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(rumors)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ToJsonByRumors
func ToJsonByRumors(rumors []*Rumor) ([]byte, error) {
	bytes, err := json.Marshal(rumors)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ToRumorsByTime
func ToRumorsByTime(txn *badger.Txn, address string) ([]*Rumor, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-rumor-time-%s", address))
	var rumors = make([]*Rumor, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		rumor, err := ToRumorByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		rumors = append(rumors, rumor)
	}
	return rumors, nil
}

// String
func (this Rumor) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal rumor", err)
		return ""
	}
	return string(bytes)
}

// UnmarshalJSON
func (this *Rumor) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["hash"] != nil {
		this.Hash = jsonMap["hash"].(string)
	}
	if jsonMap["address"] != nil {
		this.Address = jsonMap["address"].(string)
	}
	if jsonMap["transactionHash"] != nil {
		this.TransactionHash = jsonMap["transactionHash"].(string)
	}
	if jsonMap["time"] != nil {
		this.Time = int64(jsonMap["time"].(float64))
	}
	if jsonMap["signature"] != nil {
		this.Signature = jsonMap["signature"].(string)
	}
	return nil
}

// MarshalJSON
func (this Rumor) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash            string `json:"hash"`
		Address         string `json:"address"`
		TransactionHash string `json:"transactionHash"`
		Time            int64  `json:"time"`
		Signature       string `json:"signature"`
	}{
		Hash:            this.Hash,
		Address:         this.Address,
		TransactionHash: this.TransactionHash,
		Time:            this.Time,
		Signature:       this.Signature,
	})
}
