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
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

var accountInstance *Account
var accountOnce sync.Once

// GetAccount - Returns the singleton instance of the current account
func GetAccount() *Account {
	accountOnce.Do(func() {
		accountInstance = readAccountFile()
	})
	return accountInstance
}

// Account
type Account struct {
	Address    string
	PrivateKey string
	Name       string
	Balance    int64
	Updated    time.Time
	Created    time.Time
}

// type Account interface {
// 	Key()
// 	NameKey()
// 	Set(txn *badger.Txn)
// 	VerifyAddress(hash string, signature string)
// 	UnmarshalJSON(bytes []byte)
// 	MarshalJSON()
// 	String()
// }

// ToAccountFromJson -
func ToAccountFromJson(payload []byte) (*Account, error) {
	account := &Account{}
	err := json.Unmarshal(payload, account)
	if err != nil {
		return nil, err
	}
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
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
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

// Key
func (this Account) Key() string {
	return fmt.Sprintf("table-account-%s", this.Address)
}

// NameKey
func (this Account) NameKey() string {
	return fmt.Sprintf("key-account-name-%s", strings.ToLower(this.Name))
}

// Set
func (this *Account) Set(txn *badger.Txn) error {
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
		this.Balance = int64(jsonMap["balance"].(float64))
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

	return nil
}

// MarshalJSON
func (this Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Address    string    `json:"address"`
		PrivateKey string    `json:"privateKey"`
		Name       string    `json:"name"`
		Balance    int64     `json:"balance"`
		Updated    time.Time `json:"updated"`
		Created    time.Time `json:"created"`
	}{
		Address:    this.Address,
		PrivateKey: this.PrivateKey,
		Name:       this.Name,
		Balance:    this.Balance,
		Updated:    this.Updated,
		Created:    this.Created,
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

// readAccountFile -
func readAccountFile(name_optional ...string) *Account {
	name := "account.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	fileName := utils.GetDisgoDir() + string(os.PathSeparator) + name
	if !utils.Exists(fileName) {
		publicKey, privateKey := crypto.GenerateKeyPair()
		address := crypto.ToAddress(publicKey)
		account := &Account{}
		account.Address = hex.EncodeToString(address)
		account.PrivateKey = hex.EncodeToString(privateKey)
		account.Balance = 0

		// Write account.
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(account.String()), &jsonMap)
		if err != nil {
			utils.Error("unable to create account", err)
			os.Exit(1)
		}

		jsonMap["privateKey"] = account.PrivateKey

		bytes, err := json.Marshal(jsonMap)
		if err != nil {
			utils.Error("unable to create account", err)
			os.Exit(1)
		}
		writeAccountFile(bytes, name)
		return account
	}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		utils.Error("unable to read account.json", err)
		os.Exit(1)
	}
	account, err := ToAccountFromJson(bytes)
	if err != nil {
		utils.Error("unable to read account.json", err)
		os.Exit(1)
	}
	return account
}

// writeAccountFile -
func writeAccountFile(bytes []byte, name_optional ...string) {
	name := "account.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	if !utils.Exists(utils.GetDisgoDir()) {
		err := os.MkdirAll(utils.GetDisgoDir(), 0755)
		if err != nil {
			utils.Error(fmt.Sprintf("unable to create %s directory", utils.GetDisgoDir()), err)
			os.Exit(1)
		}
	}
	fileName := utils.GetDisgoDir() + string(os.PathSeparator) + name
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		utils.Error(fmt.Sprintf("unable to write %s", fileName), err)
		os.Exit(1)
	}
	fmt.Fprintf(file, string(bytes))
}
