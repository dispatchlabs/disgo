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
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"
	"github.com/patrickmn/go-cache"
	"github.com/dgraph-io/badger"
	"fmt"
	"github.com/dispatchlabs/disgo/commons/utils"
)

// var testAccountByte = []byte("{\"address\":\"99022124e110f5a9567a334a2017bdbd41c475e3\",\"privateKey\":\"abc\",\"name\":\"test\",\"balance\":1000,\"updated\":\"2018-05-09T15:04:05Z\",\"created\":\"2018-05-09T15:04:05Z\",\"nonce\":0,\"root\":\"0x0000000000000000000000000000000000000000000000000000000000000000\",\"codehash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}")
var testAccountByte = []byte("{\"address\":\"99022124e110f5a9567a334a2017bdbd41c475e3\",\"privateKey\":\"abc\",\"name\":\"test\",\"balance\":1000,\"updated\":\"2018-05-09T15:04:05Z\",\"created\":\"2018-05-09T15:04:05Z\",\"nonce\":0}")
var testAccountAddressHash = "de3a0dba79b563588b15e38909ce206eb83dd27b53150e53c858036978b23412"
var c *cache.Cache
var db *badger.DB

//init
func init()  {
	c = cache.New(CacheTTL, CacheTTL*2)
	utils.Info("opening DB...")
	opts := badger.DefaultOptions
	opts.Dir = "." + string(os.PathSeparator) + "testdb"
	opts.ValueDir = "." + string(os.PathSeparator) + "testdb"
	db, _ = badger.Open(opts)
}

//TestGetAccount
func TestGetAccount(t *testing.T) {
	// TODO: GetAccount()
	t.Skip("Should refactor away from using a singleton for testability?")
}

//TestToAccountFromJson
func TestToAccountFromJson(t *testing.T) {
	account, err := ToAccountFromJson(testAccountByte)
	if err != nil {
		t.Fatalf("ToAccountFromJson returning error: %s", err)
	}
	testAccountStruct(t, account)
}

//TestAccountCache
func TestAccountCache(t *testing.T) {
	account := &Account{}
	account.UnmarshalJSON(testAccountByte)
	//cache := config.GetTestCache()
	account.Cache(c, time.Second * 5)
	testAccount, err := ToAccountFromCache(c, account.Address)
	fmt.Print(testAccount)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(testAccount, account) == false{
		t.Error("account not equal to testAccount")
	}
}

//TestToAccountByAddress
func TestToAccountByAddress(t *testing.T) {
	// TODO: ToAccountByAddress()
	t.Skip("Need a Badger DB mock")
}

//TestToAccountByName
func TestToAccountByName(t *testing.T) {
	// TODO: ToAccountByName()
	t.Skip("Need a Badger DB mock")
}

//TestToAccountsByName
func TestToAccountsByName(t *testing.T) {
	// TODO: ToAccountByName()
	t.Skip("Need a Badger DB mock")
}

//TestAccountSet
func TestAccountSet(t *testing.T) {
	// TODO: account.PersistAndCache()
	t.Skip("Need a Badger DB mock")
}

//TestAccountUnmarshalJSON
func TestAccountUnmarshalJSON(t *testing.T) {
	account := &Account{}
	account.UnmarshalJSON(testAccountByte)
	testAccountStruct(t, account)
}

//TestAccountMarshalJSON
func TestAccountMarshalJSON(t *testing.T) {
	account := &Account{}
	account.UnmarshalJSON(testAccountByte)
	out, err := account.MarshalJSON()
	if err != nil {
		t.Fatalf("account.MarshalJSON returning error: %s", err)
	}
	if reflect.DeepEqual(out, testAccountByte) == false {
		t.Errorf("account.MarshalJSON returning invalid value.\nGot: %s\nExpected: %s", out, testAccountByte)
	}
}

//TestReadAccountFile
func TestReadAccountFile(t *testing.T) {
	name := "test.json"
	defer testCleanAccountFile(t, name)
	newAccount := readAccountFile(name)
	// - This test no longer works because of big as a function does not return null
	//existingAccount := readAccountFile(name)
	//if reflect.DeepEqual(newAccount, existingAccount) == false {
	//	t.Error("newAccount not equal to existingAccount")
	//}
	if newAccount.Address == "" {
		t.Error("newAccount.Address is empty")
	}
	if newAccount.PrivateKey == "" {
		t.Error("newAccount.PrivateKey is empty")
	}
	if newAccount.Balance.Int64() != big.NewInt(0).Int64() {
		t.Error("newAccount.Balance is not 0")
	}
	if newAccount.Created != newAccount.Updated {
		t.Error("newAccount.Created not equal to newAccount.Updated")
	}
	//if time.Time.IsZero(newAccount.Created) {
	//	t.Skip("newAccount.Created is empty")
	//}
	//if time.Time.IsZero(newAccount.Updated) {
	//	t.Skip("newAccount.Updated is empty")
	//}
}

//testCleanAccountFile
func testCleanAccountFile(t *testing.T, name_optional ...string) func() {
	name := "test.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	fileName := utils.GetConfigDir() + string(os.PathSeparator) + name
	if utils.Exists(fileName) {
		os.Remove(fileName)
	}

	return func() {
		if utils.Exists(fileName) {
			if err := os.Remove(fileName); err != nil {
				t.Errorf("Account file error: %s", err)
			}
		}
	}
}

//testAccountStruct
func testAccountStruct(t *testing.T, account *Account) {
	if account.Address != "99022124e110f5a9567a334a2017bdbd41c475e3" {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %s", "Address", account.Address)
	}
	if account.PrivateKey != "abc" {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %s", "PrivateKey", account.PrivateKey)
	}
	if account.Name != "test" {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %s", "Name", account.Name)
	}
	if account.Balance.Int64() != big.NewInt(1000).Int64() {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %d", "Balance", account.Balance)
	}
	d, _ := time.Parse(time.RFC3339, "2018-05-09T15:04:05Z")
	if account.Updated != d {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %s", "Updated", account.Updated.String())
	}
	if account.Created != d {
		t.Errorf("account.UnmarshalJSON returning invalid %s value: %s", "Created", account.Created.String())
	}
	if account.Key() != "table-account-99022124e110f5a9567a334a2017bdbd41c475e3" {
		t.Errorf("account.Key() returning invalid %s value: %s", "Key", account.Key())
	}
	if account.NameKey() != "key-account-name-test" {
		t.Errorf("account.NameKey() returning invalid %s value: %s", "NameKey", account.NameKey())
	}
	if account.String() != string(testAccountByte) {
		t.Errorf("account.String() returning invalid value.\nGot: %s\nExpected: %s", account.String(), string(testAccountByte))
	}
}
