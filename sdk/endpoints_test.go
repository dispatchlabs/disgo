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