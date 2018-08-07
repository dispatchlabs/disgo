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
	account, err := CreateAccount("Test")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", account.ToPrettyJson())
}
