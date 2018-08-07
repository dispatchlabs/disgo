package sdk

import (
	"testing"
	"fmt"
)

func TestCreateAccount(t *testing.T) {
	account, err := CreateAccount("Test")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", account.ToPrettyJson())
}
