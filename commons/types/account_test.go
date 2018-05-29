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
	"github.com/dispatchlabs/commons/utils"
	"os"
	"reflect"
	"testing"
	"time"
)

var testAccountByte = []byte("{\"address\":\"99022124e110f5a9567a334a2017bdbd41c475e3\",\"privateKey\":\"abc\",\"name\":\"test\",\"balance\":1000,\"updated\":\"2018-05-09T15:04:05Z\",\"created\":\"2018-05-09T15:04:05Z\"}")
var testAccountAddressHash = "de3a0dba79b563588b15e38909ce206eb83dd27b53150e53c858036978b23412"

func TestGetAccount(t *testing.T) {
	// TODO: GetAccount()
	t.Skip("Should refactor away from using a singleton for testability?")
}

func TestToAccountFromJson(t *testing.T) {
	account, err := ToAccountFromJson(testAccountByte)
	if err != nil {
		t.Fatalf("ToAccountFromJson returning error: %s", err)
	}
	testAccountStruct(t, account)
}

func TestToAccountByAddress(t *testing.T) {
	// TODO: ToAccountByAddress()
	t.Skip("Need a Badger DB mock")
}

func TestToAccountByName(t *testing.T) {
	// TODO: ToAccountByName()
	t.Skip("Need a Badger DB mock")
}

func TestToAccountsByName(t *testing.T) {
	// TODO: ToAccountByName()
	t.Skip("Need a Badger DB mock")
}

func TestAccountSet(t *testing.T) {
	// TODO: account.Set()
	t.Skip("Need a Badger DB mock")
}

func TestAccountUnmarshalJSON(t *testing.T) {
	account := &Account{}
	account.UnmarshalJSON(testAccountByte)
	testAccountStruct(t, account)
}

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

func TestReadAccountFile(t *testing.T) {
	name := "test.json"
	defer testCleanAccountFile(t, name)
	newAccount := readAccountFile(name)
	existingAccount := readAccountFile(name)
	if reflect.DeepEqual(newAccount, existingAccount) == false {
		t.Error("newAccount not equal to existingAccount")
	}
	if newAccount.Address == "" {
		t.Error("newAccount.Address is empty")
	}
	if newAccount.PrivateKey == "" {
		t.Error("newAccount.PrivateKey is empty")
	}
	if newAccount.Balance != 0 {
		t.Error("newAccount.Balance is not 0")
	}
	if newAccount.Created != newAccount.Updated {
		t.Error("newAccount.Created not equal to newAccount.Updated")
	}
	if time.Time.IsZero(newAccount.Created) {
		t.Skip("newAccount.Created is empty")
	}
	if time.Time.IsZero(newAccount.Updated) {
		t.Skip("newAccount.Updated is empty")
	}
}

func testCleanAccountFile(t *testing.T, name_optional ...string) func() {
	name := "test.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	fileName := utils.GetDisgoDir() + string(os.PathSeparator) + name
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
	if account.Balance != int64(1000) {
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
