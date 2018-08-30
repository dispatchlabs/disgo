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
	//toAccount, err := CreateAccount()
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//delegates, err := GetDelegates("SeedEndpoint")
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//var hash string
	//for i:=0;i<=1;i++{
	//	hash, err = TransferTokens(delegates[0], "private key", "from address" ,toAccount.Address, 5)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}
	//
	//time.Sleep(time.Second * 10)
	//
	//tx, err := GetTransaction(delegates[0],hash)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//
	//if tx.Hash != hash {
	//	fmt.Print(tx.Hash, "vs" , hash)
	//	t.Error("invalid hash")
	//}
//}


//func TestGetTransactionsReceived(t *testing.T) {
//		toAccount, err := CreateAccount()
//		if err != nil {
//			t.Error(err)
//		}
//
//		delegates, err := GetDelegates("seed endpoint")
//		if err != nil {
//			t.Error(err)
//		}
//
//		hash, err := TransferTokens(delegates[0], "private key" , "from address" ,toAccount.Address, 5)
//		if err != nil {
//			t.Error(err)
//		}
//
//		time.Sleep(time.Second * 10)
//
//		txs, err := GetTransactionsReceived(delegates[0],toAccount.Address)
//		if err != nil {
//			t.Error(err)
//		}
//
//		var failure = true
//		for _, value := range txs {
//			if value.Hash == hash{
//				failure = false
//			}
//		}
//		if failure != false{
//			t.Error("tx not found")
//		}
//}

//func TestGetTransactionsSent(t *testing.T) {
//		toAccount, err := CreateAccount()
//		if err != nil {
//			t.Error(err)
//		}
//
//		delegates, err := GetDelegates("seed endpoint")
//		if err != nil {
//			t.Error(err)
//		}
//
//		hash, err := TransferTokens(delegates[0], "private key" , "from address" ,toAccount.Address, 5)
//		if err != nil {
//			t.Error(err)
//		}
//
//		time.Sleep(time.Second * 10)
//
//		txs, err := GetTransactionsSent(delegates[0], "from address")
//		if err != nil {
//			t.Error(err)
//		}
//
//		var failure = true
//		for _, value := range txs {
//			if value.Hash == hash{
//				failure = false
//			}
//		}
//		if failure != false{
//			t.Error("tx not found")
//		}
//}