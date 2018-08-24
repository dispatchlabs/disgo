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
	"github.com/pkg/errors"
	"github.com/patrickmn/go-cache"
	"github.com/dispatchlabs/disgo/commons/helper"
)

// Transaction - The transaction info
type Transaction struct {
	Hash      string // Hash = (Type + From + To + Value + Code + Abi + Method + Params + Time)
	Type      byte
	From      string
	To        string
	Value     int64
	Code      string
	Abi       string
	Method    string
	Params    []interface{}
	Time      int64 // Milliseconds
	Signature string
	Hertz     int64   //our version of Gas
	Receipt   Receipt // Transient
	Gossip    []Rumor // Transient
	FromName  string  // Transient
	ToName    string  // Transient
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
func (this *Transaction) Cache(cache *cache.Cache, time_optional ...time.Duration) {
	TTL := TransactionTTL
	if len(time_optional) > 0 {
		TTL = time_optional[0]
	}
	cache.Set(this.Key(), this, TTL)
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

// PersistAndCache
func (this *Transaction) Set(txn *badger.Txn, cache *cache.Cache) error {
	this.Cache(cache)

	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

// GetHashBytes
func (this Transaction) GetHashBytes() crypto.HashBytes {
	return crypto.GetHashBytes(this.Hash)
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
	value, ok := cache.Get(fmt.Sprintf("table-transaction-%s", hash))
	if !ok {
		return nil, ErrNotFound
	}
	transaction := value.(*Transaction)
	return transaction, nil
}

// ToTransactions
func ToTransactions(txn *badger.Txn) ([]*Transaction, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-transaction-"))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			return nil, err
		}
		transaction, err := ToTransactionFromJson(value)
		if err != nil {
			return nil, err
		}
		transaction.setTransients(txn)
		transactions = append(transactions, transaction)
	}
	SortByTime(transactions, false)
	return transactions, nil
}

func TransactionPaging(page int,txn *badger.Txn) ([]*Transaction, error){
	var iteratorCount = 0
	var firstItem int
	pageSize := 100
	if page <= 0 {
		return nil, ErrInvalidRequest
	}else if page == 1{
		firstItem = 1
	} else{
		firstItem = (page * pageSize) - (pageSize - 1)
	}

	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-transaction-"))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		iteratorCount++
		if iteratorCount >= firstItem && iteratorCount <= (firstItem+pageSize) {
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
			transactions = append(transactions, transaction)
		}
		if iteratorCount > (firstItem+pageSize){
			break
		}
	}
	return transactions, nil //TODO: return error if empty?
}

// ToTransactionsByFromAddress
func ToTransactionsByFromAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-from-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			return nil, err
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			return nil, err
		}
		transaction.setTransients(txn)
		transactions = append(transactions, transaction)
	}
	SortByTime(transactions, false)
	return transactions, nil
}

// ToTransactionsByToAddress
func ToTransactionsByToAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-to-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			return nil, err
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			return nil, err
		}
		transaction.setTransients(txn)
		transactions = append(transactions, transaction)
	}
	SortByTime(transactions, false)
	return transactions, nil
}

// ToTransactionsByType
func ToTransactionsByType(txn *badger.Txn, tipe byte) ([]*Transaction, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-type-%d", tipe))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			return nil, err
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			return nil, err
		}
		transaction.setTransients(txn)
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionByHash
func ToTransactionByHash(txn *badger.Txn, hash string) (*Transaction, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("table-transaction-%s", hash)))
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
	transaction.setTransients(txn)
	return transaction, err
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
	transaction.setTransients(txn)
	return transaction, err
}

// NewTransferTokensTransaction -
func NewTransferTokensTransaction(privateKey string, from, to string, value int64, hertz int64, timeInMiliseconds int64) (*Transaction, error) {
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeTransferTokens
	transaction.From = from
	transaction.To = to
	transaction.Value = value
	transaction.Time, err = checkTime(timeInMiliseconds)
	if err != nil {
		return nil, err
	}
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewDeployContractTransaction -
func NewDeployContractTransaction(privateKey string, from string, code string, abi string, timeInMiliseconds int64) (*Transaction, error) {
	if abi == "" {
		return nil, errors.Errorf("cannot have empty abi")
	}
	if code == "" {
		return nil, errors.Errorf("cannot have empty code")
	}
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeDeploySmartContract
	transaction.From = from
	transaction.To = ""
	transaction.Code = code
	transaction.Abi = abi
	transaction.Time, err = checkTime(timeInMiliseconds)
	if err != nil {
		return nil, err
	}
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewExecuteContractTransaction -
func NewExecuteContractTransaction(privateKey string, from string, to string, abi string, method string, params []interface{}, timeInMiliseconds int64) (*Transaction, error) {
	if abi == "" {
		return nil, errors.Errorf("cannot have empty abi")
	}
	if method == "" {
		return nil, errors.Errorf("cannot have empty method")
	}
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeExecuteSmartContract
	transaction.From = from
	transaction.To = to
	transaction.Abi = abi
	transaction.Method = method
	transaction.Params = params
	transaction.Time, err = checkTime(timeInMiliseconds)
	if err != nil {
		return nil, err
	}
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewHash
func (this Transaction) NewHash() (string, error) {
	fromBytes, err := hex.DecodeString(this.From)
	if err != nil {
		utils.Error("unable decode from", err)
		return "", err
	}
	toBytes, err := hex.DecodeString(this.To)
	if err != nil {
		utils.Error("unable decode to", err)
		return "", err
	}
	codeBytes, err := hex.DecodeString(this.Code)
	if err != nil {
		utils.Error("unable decode code", err)
		return "", err
	}
	var values = []interface{}{
		this.Type,
		fromBytes,
		toBytes,
		this.Value,
		codeBytes,
		[]byte(this.Abi),
		[]byte(this.Method),
		// TODO: this.Params,
		this.Time,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Fatal("unable to write transaction bytes to buffer", err)
			return "", err
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	return hex.EncodeToString(hash[:]), nil
}

// NewSignature
func (this Transaction) NewSignature(privateKey string) (string, error) {
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
func (this Transaction) Verify() error {
	if len(this.Hash) != crypto.HashLength*2 {
		return errors.New("invalid hash")
	}
	if len(this.From) != crypto.AddressLength*2 {
		return errors.New("invalid from address")
	}
	if len(this.Signature) != crypto.SignatureLength*2 {
		return errors.New("invalid signature")
	}
	if this.From == this.To {
		return errors.New("from address cannot equal to address")
	}

	// Type?
	switch this.Type {
	case TypeTransferTokens:
		if len(this.To) != crypto.AddressLength*2 {
			return errors.New("invalid to address")
		}
		if this.Value <= 0 {
			return errors.New("value cannot be less than or equal to zero")
		}
		break
	case TypeDeploySmartContract:
		if len(this.To) != 0 {
			return errors.New("to address must be blank for a deployment of a smart contract")
		}
		if len(this.Code) == 0 {
			return errors.New("invalid code")
		}
		if len(this.Abi) == 0 {
			return errors.New("invalid abi")
		}
		break
	case TypeExecuteSmartContract:
		if len(this.To) != crypto.AddressLength*2 {
			return errors.New("invalid to address")
		}

		// TODO: Should we check method?
		break
	}

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
	if address != this.From {
		return errors.New("from address does not match the computed address from hash and signature")
	}
	if !crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes) {
		return errors.New("invalid signature")
	}

	return nil
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

// String
func (this Transaction) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal transaction", err)
		return ""
	}
	return string(bytes)
}

// UnmarshalJSON
func (this *Transaction) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	var ok bool
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["hash"] != nil {
		this.Hash, ok = jsonMap["hash"].(string)
		if !ok {
			return errors.Errorf("value for field 'hash' must be a string")
		}
	}
	if jsonMap["type"] != nil {
		typ, ok := jsonMap["type"].(float64)
		if !ok {
			return errors.Errorf("value for field 'type' must be a number")
		}
		this.Type = byte(typ)
	}
	if jsonMap["from"] != nil {
		this.From, ok = jsonMap["from"].(string)
		if !ok {
			return errors.Errorf("value for field 'from' must be a string")
		}
	}
	if jsonMap["to"] != nil {
		this.To, ok = jsonMap["to"].(string)
		if !ok {
			return errors.Errorf("value for field 'to' must be a string")
		}
	}
	if jsonMap["value"] != nil {
		value, ok := jsonMap["value"].(float64)
		if !ok {
			return errors.Errorf("value for field 'value' must be a number")
		}
		this.Value = int64(value)

	}
	if jsonMap["code"] != nil {
		this.Code, ok = jsonMap["code"].(string)
		if !ok {
			return errors.Errorf("value for field 'code' must be a string")
		}
	}
	if jsonMap["abi"] != nil {
		this.Abi, ok = jsonMap["abi"].(string)
		if !ok {
			return errors.Errorf("value for field 'abi' must be a string")
		}
		to, _ := jsonMap["to"].(string)
		method, _ := jsonMap["method"].(string)
		if len(to) > 0 && len(method) > 0 && this.Abi == "" {
			return errors.Errorf("value for field 'abi' is invalid")
		}
	}
	if jsonMap["method"] != nil {
		this.Method, ok = jsonMap["method"].(string)
		if !ok {
			return errors.Errorf("value for field 'method' must be a string")
		}
	}
	if jsonMap["params"] != nil {
		params, err := helper.GetConvertedParams(jsonMap)
		if err != nil {
			return err
		}
		this.Params = params
	}
	if jsonMap["time"] != nil {
		t, ok := jsonMap["time"].(float64)
		if !ok {
			return errors.Errorf("value for field 'time' must be a number")
		}
		txTime, err := checkTime(int64(t))
		if err != nil {
			return err
		}
		this.Time = txTime
	}
	if jsonMap["signature"] != nil {
		this.Signature, ok = jsonMap["signature"].(string)
		if !ok {
			return errors.Errorf("value for field 'signature' must be a string")
		}
	}
	if jsonMap["hertz"] != nil {
		hertz, ok := jsonMap["type"].(float64)
		if !ok {
			return errors.Errorf("value for field 'hertz' must be a number")
		}
		this.Hertz = int64(hertz)
	}
	return nil
}

// MarshalJSON
func (this Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash      string        `json:"hash"`
		Type      byte          `json:"type"`
		From      string        `json:"from"`
		To        string        `json:"to,omitempty"`
		Value     int64         `json:"value,omitempty"`
		Code      string        `json:"code,omitempty"`
		Abi       string        `json:"abi,omitempty"`
		Method    string        `json:"method,omitempty"`
		Params    []interface{} `json:"params,omitempty"`
		Time      int64         `json:"time"`
		Signature string        `json:"signature"`
		Hertz     int64         `json:"hertz"`
		Receipt   Receipt       `json:"receipt,omitempty"`
		Gossip    []Rumor       `json:"gossip,omitempty"`
		FromName  string        `json:"fromName,omitempty"`
		ToName    string        `json:"toName,omitempty"`
	}{
		Hash:      this.Hash,
		Type:      this.Type,
		From:      this.From,
		To:        this.To,
		Value:     this.Value,
		Code:      this.Code,
		Abi:       this.Abi,
		Method:    this.Method,
		Params:    this.Params,
		Time:      this.Time,
		Signature: this.Signature,
		Hertz:     this.Hertz,
		Receipt:   this.Receipt,
		Gossip:    this.Gossip,
		FromName:  this.FromName,
		ToName:    this.ToName,
	})
}

// Equals
func (this Transaction) Equals(other string) bool {
	return this.Hash == other
}

func checkTime(txTime int64) (int64, error) {
	now := utils.ToMilliSeconds(time.Now())
	if now + TxReceiveWiggle < txTime { // Adding "wiggle room" to allow for clock variances
		return txTime, errors.Errorf("transaction time cannot be in the future")
	} else if txTime < 0 {
		return txTime, errors.Errorf("transaction time cannot be negative")
	}
	//TODO: need to have a limit check here that it is not older than some value whether that is static at startup or relative to current time.
	//TODO: Talking with Avery, should be related to page TS limits.  This will not be the appropriate place for the check but will suffice for the moment.

	return txTime, nil
}

// setTransients
func (this *Transaction) setTransients(txn *badger.Txn) {
	fromAccount, err := ToAccountByAddress(txn, this.From)
	if err == nil {
		this.FromName = fromAccount.Name
	}
	toAccount, err := ToAccountByAddress(txn, this.To)
	if err == nil {
		this.ToName = toAccount.Name
	}
	receipt, err := ToReceiptFromTransactionHash(txn, this.Hash)
	if err == nil {
		this.Receipt = *receipt
	}
	gossip, err := ToGossipByTransactionHash(txn, this.Hash)
	if err == nil {
		this.Gossip = gossip.Rumors
	}
}
