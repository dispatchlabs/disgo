package sdk

import (
	"testing"
	"fmt"
)

func TestCreateAccount(t *testing.T) {
	// Testing WITHOUT name
	account, err := CreateAccount()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", account.ToPrettyJson())

	// Testing WITH name
	account, err = CreateAccount("Test")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", account.ToPrettyJson())
}

func TestGetDelegates(t *testing.T) {
	// Testing WITHOUT seedUrl
	delegates, err := GetDelegates()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", delegates)

	// Testing WITH seedUrl
	delegates, err = GetDelegates("seed.dispatchlabs.io:1975")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", delegates)
}

func TestGetTransaction(t *testing.T) {
	toAccount, err := CreateAccount()
	if err != nil {
		t.Error(err)
	}

	delegates, err := GetDelegates("127.0.0.1:3500")
	if err != nil {
		t.Error(err)
	}

	hash, err := TransferTokens(delegates[0], "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a", "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c", toAccount.Address, 5)
	if err != nil {
		t.Error(err)
	}

	tx, err := GetTransaction(delegates[0],hash)
	if err != nil {
		t.Error(err)
	}

	if tx.Hash != hash {
		fmt.Print(tx.Hash, "vs" , hash)
		t.Error("invalid hash")
	}
}