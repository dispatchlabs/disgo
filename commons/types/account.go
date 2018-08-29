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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/patrickmn/go-cache"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"math/big"
)

var accountInstance *Account
var accountOnce sync.Once

// Account
type Account struct {
	Address         string
	PrivateKey      string
	Name            string
	Balance         *big.Int
	TransactionHash string // Smart contract
	Updated         time.Time
	Created         time.Time

	// From Ethereum Account
	Nonce    uint64
	Root     crypto.HashBytes // merkle root of the storage trie
	CodeHash []byte
}

// Key
func (this Account) Key() string {
	return fmt.Sprintf("table-account-%s", this.Address)
}

// NameKey
func (this Account) NameKey() string {
	return fmt.Sprintf("key-account-name-%s", strings.ToLower(this.Name))
}

//Cache
func (this *Account) Cache(cache *cache.Cache, time_optional ...time.Duration) {
	TTL := AccountTTL
	if len(time_optional) > 0 {
		TTL = time_optional[0]
	}
	cache.Set(this.Key(), this, TTL)
}

//Persist
func (this *Account) Persist(txn *badger.Txn) error {
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.NameKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// PersistAndCache
func (this *Account) Set(txn *badger.Txn, cache *cache.Cache) error {
	this.Cache(cache)
	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON
func (this *Account) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}
	if jsonMap["address"] != nil {
		this.Address = jsonMap["address"].(string)
	}
	if jsonMap["privateKey"] != nil {
		this.PrivateKey = jsonMap["privateKey"].(string)
	}
	if jsonMap["name"] != nil {
		this.Name = jsonMap["name"].(string)
	}
	if jsonMap["balance"] != nil {
		this.Balance = big.NewInt(int64(jsonMap["balance"].(float64)))
	}
	if jsonMap["transactionHash"] != nil {
		this.TransactionHash = jsonMap["transactionHash"].(string)
	}
	if jsonMap["updated"] != nil {
		updated, err := time.Parse(time.RFC3339, jsonMap["updated"].(string))
		if err != nil {
			return err
		}
		this.Updated = updated
	}
	if jsonMap["created"] != nil {
		created, err := time.Parse(time.RFC3339, jsonMap["created"].(string))
		if err != nil {
			return err
		}
		this.Created = created
	}
	if jsonMap["nonce"] != nil {
		this.Nonce = uint64(jsonMap["nonce"].(float64))
	}
	// if jsonMap["root"] != nil {
	// 	this.Root = crypto.GetHashBytes(jsonMap["root"].(string))
	// }
	// if jsonMap["codehash"] != nil {
	// 	this.CodeHash = crypto.GetHashBytes(jsonMap["codehash"].(string)).Bytes()
	// }

	return nil
}

// MarshalJSON
func (this Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Address         string    `json:"address"`
		PrivateKey      string    `json:"privateKey,omitempty"`
		Name            string    `json:"name"`
		Balance         int64     `json:"balance"`
		TransactionHash string    `json:"transactionHash,omitempty"`
		Updated         time.Time `json:"updated"`
		Created         time.Time `json:"created"`
		Nonce           uint64    `json:"nonce"`
		// Root       string    `json:"root"`
		// CodeHash   string    `json:"codehash"`
	}{
		Address:         this.Address,
		PrivateKey:      this.PrivateKey,
		Name:            this.Name,
		Balance:         this.Balance.Int64(),
		TransactionHash: this.TransactionHash,
		Updated:         this.Updated,
		Created:         this.Created,
		Nonce:           this.Nonce,
		// Root:       crypto.Encode(this.Root.Bytes()),
		// CodeHash:   crypto.Encode(this.CodeHash),
	})
}

// String
func (this Account) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal account", err)
		return ""
	}
	return string(bytes)
}

func (this Account) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal transaction", err)
		return ""
	}
	return string(bytes)
}


// GetAccount - Returns the singleton instance of the current account
func GetAccount() *Account {
	accountOnce.Do(func() {
		accountInstance = readAccountFile()
	})
	return accountInstance
}

// ToAccountFromJson -
func ToAccountFromJson(payload []byte) (*Account, error) {
	account := &Account{}
	err := json.Unmarshal(payload, account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// ToAccountFromCache -
func ToAccountFromCache(cache *cache.Cache, address string) (*Account, error) {
	value, ok := cache.Get(fmt.Sprintf("table-account-%s", address))
	if !ok {
		return nil, ErrNotFound
	}
	account := value.(*Account)
	return account, nil
}

// ToAccountByAddress
func ToAccountByAddress(txn *badger.Txn, address string) (*Account, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("table-account-%s", address)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	account, err := ToAccountFromJson(value)
	if err != nil {
		return nil, err
	}
	return account, err
}

// ToAccountByName
func ToAccountByName(txn *badger.Txn, name string) (*Account, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("key-account-name-%s", name)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	account, err := ToAccountByAddress(txn, string(value))
	if err != nil {
		return nil, err
	}
	return account, err
}

// ToAccountsByName
func ToAccountsByName(name string, txn *badger.Txn) ([]*Account, error) {
	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-account-name-%s", name))
	var Accounts = make([]*Account, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		Account, err := ToAccountByAddress(txn, string(value))
		if err != nil {
			utils.Error(err)
			continue
		}
		Accounts = append(Accounts, Account)
	}
	return Accounts, nil
}
					//txn, start, pageNumber, pageSize
func AccountPaging(txn *badger.Txn, startingHash string, page, pageSize int) ([]*Account, error){
	var iteratorCount = 0
	var firstItem int
	if pageSize <= 0 || pageSize > 100{
		return nil, ErrInvalidRequestPageSize
	}
	if page <= 0 {
		return nil, ErrInvalidRequestPage
	}else if page == 1{
		firstItem = 1
	} else{
		firstItem = (page * pageSize) - (pageSize - 1)
	}
	var item []byte
	prefix := []byte(fmt.Sprintf("table-account-"))
	if startingHash != "" {
		thing, err := ToAccountByAddress(txn,startingHash)
		if err != nil {
			return nil, ErrInvalidRequestHash
		}
		item = []byte(thing.Key())
	} else{
		item = prefix
	}

	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	var Accounts = make([]*Account, 0)
	for iterator.Seek(item); iterator.ValidForPrefix(prefix); iterator.Next() {
		iteratorCount++
		if iteratorCount >= firstItem && iteratorCount < (firstItem+pageSize) {
			item := iterator.Item()
			value, err := item.Value()
			if err != nil {
				utils.Error(err)
				continue
			}
			Account, err := ToAccountFromJson(value)
			if err != nil {
				utils.Error(err)
				continue
			}
			Accounts = append(Accounts, Account)
		}
		if iteratorCount >= (firstItem+pageSize){
			break
		}
	}
	return Accounts, nil //TODO: return error if empty?
}

// readAccountFile -
func readAccountFile(name_optional ...string) *Account {
	name := "account.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	fileName := utils.GetConfigDir() + string(os.PathSeparator) + name
	if !utils.Exists(fileName) {
		publicKey, privateKey := crypto.GenerateKeyPair()
		address := crypto.ToAddress(publicKey)
		account := &Account{}
		account.Address = hex.EncodeToString(address)
		account.PrivateKey = hex.EncodeToString(privateKey)
		account.Balance = big.NewInt(0)
		account.Name = ""
		now := time.Now()
		account.Created = now
		account.Updated = now

		// Write account.
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(account.String()), &jsonMap)
		if err != nil {
			utils.Fatal("unable to create account", err)
		}

		jsonMap["privateKey"] = account.PrivateKey

		bytes, err := json.Marshal(jsonMap)
		if err != nil {
			utils.Fatal("unable to create account", err)
		}
		writeAccountFile(bytes, name)
	}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		utils.Fatal("unable to read account.json", err)
	}
	account, err := ToAccountFromJson(bytes)
	if err != nil {
		utils.Fatal("unable to read account.json", err)
	}
	return account
}

// writeAccountFile -
func writeAccountFile(bytes []byte, name_optional ...string) {
	name := "account.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	if !utils.Exists(utils.GetConfigDir()) {
		err := os.MkdirAll(utils.GetConfigDir(), 0755)
		if err != nil {
			utils.Fatal(fmt.Sprintf("unable to create %s directory", utils.GetConfigDir()), err)
		}
	}
	fileName := utils.GetConfigDir() + string(os.PathSeparator) + name
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		utils.Fatal(fmt.Sprintf("unable to write %s", fileName), err)
	}
	fmt.Fprintf(file, string(bytes))
}
