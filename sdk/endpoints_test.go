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

// func TestGetDelegates(t *testing.T) {
// 	// Testing WITHOUT seedUrl
// 	delegates, err := GetDelegates()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Printf("%v\n", delegates)

// 	// Testing WITH seedUrl
// 	delegates, err = GetDelegates("seed.dispatchlabs.io:1975")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Printf("%v\n", delegates)
// }

//func TestGetTransaction(t *testing.T) {
//	toAccount, err := CreateAccount()
//	if err != nil {
//		t.Error(err)
//	}
//
//	delegates, err := GetDelegates("127.0.0.1:3500")
//	if err != nil {
//		t.Error(err)
//	}
//
//	//hash, err := TransferTokens(delegates[0], toAccount.Address, 5)
//	if err != nil {
//		t.Error(err)
//	}
//
//	tx, err := GetTransaction(delegates[0],hash)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if tx.Hash != hash {
//		fmt.Print(tx.Hash, "vs" , hash)
//		t.Error("invalid hash")
//	}
//}