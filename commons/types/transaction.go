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
	"time"

	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
)

// Types
const (
	TransactionTypeTransferTokens = 0
	TransactionTypeSetName        = 1
	TransactionTypeSmartContract  = 2
)

// Transaction - The transaction info
type Transaction struct {
	Hash      string // Hash = (Type + From + To + Value + Code + Method + Time)
	Type      byte
	From      string
	To        string
	Value     int64
	Code      string
	Method    string
	Time      int64 // Milliseconds
	Signature string
	Hertz     int64  //our version of Gas
	FromName  string // Transient
	ToName    string // Transient
}

// Key
func (this Transaction) Key() string {
	return fmt.Sprintf("table-transaction-%s", this.Hash)
}

// TypeKey
func (this Transaction) TypeKey() string {
	return fmt.Sprintf("key-transaction-type-%d-%d-%s", this.Type, this.Time, this.Hash)
}

// TimeKey
func (this Transaction) TimeKey() string {
	return fmt.Sprintf("key-transaction-time-%d-%s", this.Time, this.Hash)
}

// FromKey
func (this Transaction) FromKey() string {
	return fmt.Sprintf("key-transaction-from-%s-%d", this.From, this.Time)
}

// ToKey
func (this Transaction) ToKey() string {
	return fmt.Sprintf("key-transaction-to-%s-%d", this.To, this.Time)
}

//Cache
func (this *Transaction) Cache(cache *cache.Cache){
	cache.Set(this.Hash, this, TransactionTTL)
}

// Persist
func (this *Transaction) Persist(txn *badger.Txn) error {
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.TypeKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.TimeKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.FromKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.ToKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// Set
func (this *Transaction) Set(txn *badger.Txn,cache *cache.Cache) error {
	this.Cache(cache)

	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}


// UnmarshalJSON
func (this *Transaction) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["hash"] != nil {
		this.Hash = jsonMap["hash"].(string)
	}
	if jsonMap["type"] != nil {
		this.Type = byte(jsonMap["type"].(float64))
	}
	if jsonMap["from"] != nil {
		this.From = jsonMap["from"].(string)
	}
	if jsonMap["to"] != nil {
		this.To = jsonMap["to"].(string)
	}
	if jsonMap["value"] != nil {
		this.Value = int64(jsonMap["value"].(float64))
	}
	if jsonMap["time"] != nil {
		this.Time = int64(jsonMap["time"].(float64))
	}
	if jsonMap["signature"] != nil {
		this.Signature = jsonMap["signature"].(string)
	}
	if jsonMap["hertz"] != nil {
		this.Hertz = int64(jsonMap["hertz"].(float64))
	}
	if jsonMap["fromName"] != nil {
		this.FromName = jsonMap["fromName"].(string)
	}
	if jsonMap["toName"] != nil {
		this.ToName = jsonMap["toName"].(string)
	}

	if jsonMap["code"] != nil {
		this.Code = jsonMap["code"].(string)
	}
	if jsonMap["method"] != nil {
		this.Method = jsonMap["method"].(string)
	}

	return nil
}

// MarshalJSON
func (this Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash      string `json:"hash"`
		Type      byte   `json:"type"`
		From      string `json:"from"`
		To        string `json:"to"`
		Value     int64  `json:"value"`
		Code      string `json:"code"`
		Method    string `json:"method"`
		Time      int64  `json:"time"`
		Signature string `json:"signature"`
		Hertz     int64  `json:"hertz"`
		FromName  string `json:"fromName"`
		ToName    string `json:"toName"`
	}{
		Hash:      this.Hash,
		Type:      this.Type,
		From:      this.From,
		To:        this.To,
		Value:     this.Value,
		Code:      this.Code,
		Method:    this.Method,
		Time:      this.Time,
		Signature: this.Signature,
		Hertz:     this.Hertz,
		FromName:  this.FromName,
		ToName:    this.ToName,
	})
}

// String
func (this Transaction) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal transaction", err)
		return ""
	}
	return string(bytes)
}

// GetHashBytes
func (this Transaction) GetHashBytes() crypto.HashBytes {
	return crypto.GetHashBytes(this.Hash)
}

// NewHash
func (this Transaction) NewHash() string {
	fromBytes, err := hex.DecodeString(this.From)
	if err != nil {
		utils.Error("unable toBytes decode from", err)
		return ""
	}
	toBytes, err := hex.DecodeString(this.To)
	if err != nil {
		utils.Error("unable toBytes decode to", err)
		return ""
	}
	codeBytes, err := hex.DecodeString(this.Code)
	if err != nil {
		utils.Error("unable toBytes decode data", err)
		return ""
	}
	var values = []interface{}{
		this.Type,
		fromBytes,
		toBytes,
		this.Value,
		codeBytes,
		this.Time,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Fatal("unable to write transaction bytes to buffer", err)
			return ""
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	return hex.EncodeToString(hash[:])
}

// Verify
func (this Transaction) Verify() bool {
	if len(this.Hash) != crypto.HashLength*2 {
		utils.Debug("Invalid Hash")
		return false
	}
	if len(this.From) != crypto.AddressLength*2 {
		utils.Debug("Invalid From Address")
		return false
	}
	if this.To != "" && len(this.To) != crypto.AddressLength*2 {
		utils.Debug("Invalid To Address")
		return false
	}
	if this.To != "" && this.Value < 0 {
		return false
	}
	if len(this.Signature) != crypto.SignatureLength*2 {
		utils.Debug("Invalid Signature")
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

	// Derived address from publicKeyBytes match from?
	address := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if address != this.From {
		return false
	}
	return crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes)
}

// ToTime
func (this Transaction) ToTime() time.Time {
	return time.Unix(0, this.Time*int64(time.Millisecond))
}

// CalculateHash (MerkleTree)
func (this Transaction) CalculateHash() []byte {
	from, err := hex.DecodeString(this.From)
	if err != nil {
		utils.Fatal("unable to decode from", err)
		panic(err)
	}
	to, err := hex.DecodeString(this.To)
	if err != nil {
		utils.Fatal("unable to decode to", err)
		panic(err)
	}
	signature, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Fatal("unable to decode signature", err)
		panic(err)
	}
	var values = []interface{}{
		this.Type,
		from,
		to,
		this.Value,
		this.Time,
		signature,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.BigEndian, value)
		if err != nil {
			utils.Fatal("unable to write transaction bytes to buffer", err)
			panic(err)
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	return hash[:]
}

// Equals
func (this Transaction) Equals(other string) bool {
	return this.Hash == other
}


// ToTransactionFromJson -
func ToTransactionFromJson(payload []byte) (*Transaction, error) {
	transaction := &Transaction{}
	err := json.Unmarshal(payload, transaction)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// ToTransactionFromCache -
func ToTransactionFromCache(cache *cache.Cache, hash string) (*Transaction, error) {
	value, ok :=cache.Get(hash)
	if !ok{
		return nil, ErrNotFound
	}
	transaction := value.(*Transaction)
	return transaction, nil
}

// ToTransactions
func ToTransactions(txn *badger.Txn) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-transaction-"))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionFromJson(value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByFromAddress
func ToTransactionsByFromAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-from-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByToAddress
func ToTransactionsByToAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-to-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByType
func ToTransactionsByType(txn *badger.Txn, tipe byte) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-type-%d", tipe))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionByKey
func ToTransactionByKey(txn *badger.Txn, key []byte) (*Transaction, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	transaction, err := ToTransactionFromJson(value)
	if err != nil {
		return nil, err
	}
	return transaction, err
}

// NewTransaction -
func NewTransaction(privateKey string, tipe byte, from, to string, value, hertz, theTime int64) (*Transaction, error) {
	transaction := &Transaction{}
	transaction.Type = tipe
	transaction.From = from
	transaction.To = to
	transaction.Value = value
	transaction.Time = theTime

	return setTxHashAndSignature(transaction, privateKey)
}

// NewContractTransaction -
func NewContractTransaction(privateKey string, from string, code string, timeInMiliseconds int64) (*Transaction, error) {
	transaction := &Transaction{}
	transaction.From = from
	transaction.Code = code
	transaction.Time = timeInMiliseconds

	return setTxHashAndSignature(transaction, privateKey)
}

// NewContractCallTransaction -
func NewContractCallTransaction(privateKey string, from string, to string, code string, timeInMiliseconds int64, method string, value int64) (*Transaction, error) {
	transaction := &Transaction{}
	transaction.From = from
	transaction.To = to
	transaction.Code = code
	transaction.Time = timeInMiliseconds
	transaction.Method = method
	transaction.Value = value

	return setTxHashAndSignature(transaction, privateKey)
}

func setTxHashAndSignature(tx *Transaction, privateKey string) (*Transaction, error) {
	tx.Hash = tx.NewHash()

	hashBytes, err := hex.DecodeString(tx.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return nil, err
	}

	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		utils.Error("unable to decode privateKey", err)
		return nil, err
	}

	signatureBytes, err := crypto.NewSignature(privateKeyBytes, hashBytes)
	if err != nil {
		return nil, err
	}

	tx.Signature = hex.EncodeToString(signatureBytes)

	return tx, nil
}

