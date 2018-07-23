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
	"fmt"
	"github.com/dispatchlabs/disgo/commons/utils"
	"testing"
	"time"
	"reflect"
)

//testMockTransaction
func testMockTransaction(t *testing.T) *Transaction {
	//codeBytes := make([]byte, 0)
	d, _ := time.Parse(time.RFC3339, "2018-07-09T15:04:05Z")
	//privKeyBytes, err := hex.DecodeString("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a")
	//if err != nil {
	//	t.Fatalf("Could not create privKeyBytes %s", err.Error())
	//	return nil
	//} else {
	tx, err := NewTransferTokensTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
		"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c",
		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
		1,
		0,
		utils.ToMilliSeconds(d),
	)

	if err != nil {
		t.Fatalf("Could not create transaction %s", err.Error())
	}
	return tx
	//}
}

//TestTransactionCache
func TestTransactionCache(t *testing.T) {
	tx := testMockTransaction(t)
	tx.Cache(c, time.Second * 5)
	testTx, err := ToTransactionFromCache(c, tx.Hash)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(testTx, tx) == false{
		t.Error("tx not equal to testTx")
	}
}

// TestTransactionVerify
func TestTransactionVerify(t *testing.T) {
	tx := testMockTransaction(t)

	err := tx.Verify()
	if err == nil {
		t.Log("transaction verified")
	} else {
		t.Error("cannot verify transaction", err)
	}
}

//TestNewTransaction
func TestNewTransaction(t *testing.T) {
	tx := testMockTransaction(t)
	if tx == nil {
		t.Error("Unable to create Transaction")
	}
	if tx.Signature == "" {
		t.Error("Unable to create Signature on Transaction")
	}
}

//TestNewHash
func TestNewHash(t *testing.T) {
	tx := testMockTransaction(t)
	hash, _ := tx.NewHash()

	if hash == "" {
		t.Error("unable to create new hash for a TX")
	}

	if len(hash) != 64 {
		t.Error("hash length is NOT valid")
	}
}

//TestBadKeyTransaction
func TestPrintTransaction(t *testing.T) {
	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var theTime = utils.ToMilliSeconds(time.Now())
	//var method = "getVar5"
	//var params = make([]interface{}, 0)

	var tx, _ = NewTransferTokensTransaction(
		privateKey,
		from,
		"d5765c93699c96327753230ac3d78edb3b34236b",
		1,
		1,
		theTime,
	)
	fmt.Printf("EXECUTE_Get: \n\n%s\n\n", tx.String())
}

func TestPrintNewDeployTx(t *testing.T) {
	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var code = "608060405234801561001057600080fd5b506040805190810160405280600d81526020017f61616161616161616161616161000000000000000000000000000000000000008152506000908051906020019061005c9291906100f8565b5060006002600001819055506000600260010160006101000a81548160ff0219169083151502179055506001600260010160016101000a81548160ff021916908360ff1602179055506040805190810160405280600b81526020017f62626262626262626262620000000000000000000000000000000000000000008152506002800190805190602001906100f29291906100f8565b5061019d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013957805160ff1916838001178555610167565b82800160010185558215610167579182015b8281111561016657825182559160200191906001019061014b565b5b5090506101749190610178565b5090565b61019a91905b8082111561019657600081600090555060010161017e565b5090565b90565b6109c5806101ac6000396000f300608060405260043610610099576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806333e538e91461009e57806334e45f531461012e5780633a458b1f146101975780636e59c66c1461024657806378d8866e146102f557806379af647314610385578063cb69e3001461039c578063e4e38c7c14610405578063e98483cb14610495575b600080fd5b3480156100aa57600080fd5b506100b3610525565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100f35780820151818401526020810190506100d8565b50505050905090810190601f1680156101205780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561013a57600080fd5b50610195600480360381019080803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506105c7565b005b3480156101a357600080fd5b506101ac6105e3565b60405180858152602001841515151581526020018360ff1660ff16815260200180602001828103825283818151815260200191508051906020019080838360005b838110156102085780820151818401526020810190506101ed565b50505050905090810190601f1680156102355780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b34801561025257600080fd5b506102f3600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506106b3565b005b34801561030157600080fd5b5061030a6106e5565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561034a57808201518184015260208101905061032f565b50505050905090810190601f1680156103775780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561039157600080fd5b5061039a610783565b005b3480156103a857600080fd5b50610403600480360381019080803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919291929050505061079a565b005b34801561041157600080fd5b5061041a6107b4565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561045a57808201518184015260208101905061043f565b50505050905090810190601f1680156104875780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156104a157600080fd5b506104aa610856565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156104ea5780820151818401526020810190506104cf565b50505050905090810190601f1680156105175780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b606060008054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156105bd5780601f10610592576101008083540402835291602001916105bd565b820191906000526020600020905b8154815290600101906020018083116105a057829003601f168201915b5050505050905090565b806002800190805190602001906105df9291906108f4565b5050565b60028060000154908060010160009054906101000a900460ff16908060010160019054906101000a900460ff1690806002018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156106a95780601f1061067e576101008083540402835291602001916106a9565b820191906000526020600020905b81548152906001019060200180831161068c57829003601f168201915b5050505050905084565b81600090805190602001906106c99291906108f4565b5080600190805190602001906106e09291906108f4565b505050565b60008054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561077b5780601f106107505761010080835404028352916020019161077b565b820191906000526020600020905b81548152906001019060200180831161075e57829003601f168201915b505050505081565b600260000160008154809291906001019190505550565b80600090805190602001906107b09291906108f4565b5050565b606060018054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561084c5780601f106108215761010080835404028352916020019161084c565b820191906000526020600020905b81548152906001019060200180831161082f57829003601f168201915b5050505050905090565b60018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156108ec5780601f106108c1576101008083540402835291602001916108ec565b820191906000526020600020905b8154815290600101906020018083116108cf57829003601f168201915b505050505081565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061093557805160ff1916838001178555610963565b82800160010185558215610963579182015b82811115610962578251825591602001919060010190610947565b5b5090506109709190610974565b5090565b61099691905b8082111561099257600081600090555060010161097a565b5090565b905600a165627a7a72305820074899e01fcd4d2ae6ffd88a31c3bc77477fff7ed19e4bf8dc4af234d33dd4b80029"
	var abi = `[
	{
		"constant": true,
		"inputs": [],
		"name": "getVar5",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value",
				"type": "string"
			}
		],
		"name": "setVar6Var4",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var6",
		"outputs": [
			{
				"name": "var1",
				"type": "uint256"
			},
			{
				"name": "var2",
				"type": "bool"
			},
			{
				"name": "var3",
				"type": "uint8"
			},
			{
				"name": "var4",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value1",
				"type": "string"
			},
			{
				"name": "value2",
				"type": "string"
			}
		],
		"name": "setMultiple",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var5",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "incVar6Var1",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value",
				"type": "string"
			}
		],
		"name": "setVar5",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "getVar55",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var55",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	}
]`
	var theTime = utils.ToMilliSeconds(time.Now())

	var tx, _ = NewDeployContractTransaction(
		privateKey,
		from,
		code,
		abi,
		theTime,
	)

	fmt.Printf("DEPLOY: %s", tx.String())
}

func TestPrintNewExecuteTx(t *testing.T) {
	// Taken from Genesis
	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var to = "10412d6de794ab228e735eb0622f2deffca2edc5" // "c3be1a3a5c6134cca51896fadf032c4c61bc355e"
	var abi = `[
	{
		"constant": true,
		"inputs": [],
		"name": "getVar5",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value",
				"type": "string"
			}
		],
		"name": "setVar6Var4",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var6",
		"outputs": [
			{
				"name": "var1",
				"type": "uint256"
			},
			{
				"name": "var2",
				"type": "bool"
			},
			{
				"name": "var3",
				"type": "uint8"
			},
			{
				"name": "var4",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value1",
				"type": "string"
			},
			{
				"name": "value2",
				"type": "string"
			}
		],
		"name": "setMultiple",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var5",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "incVar6Var1",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "value",
				"type": "string"
			}
		],
		"name": "setVar5",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "getVar55",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "var55",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	}
]`

	var theTime = utils.ToMilliSeconds(time.Now())
	var method = "setMultiple"
	var params = make([]interface{}, 1)
	params[0] = "5555"

	var tx, _ = NewExecuteContractTransaction(
		privateKey,
		from,
		to,
		hex.EncodeToString([]byte(abi)),
		method,
		params,
		theTime,
	)
	fmt.Printf("DEPLOY: %s", tx.String())

}

//func TestPrintTransaction3(t *testing.T) {
//	var privateKey= "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
//	var from= "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
//	var theTime= utils.ToMilliSeconds(time.Now())
//	var method = "getVar5"
//	var params = make([]interface{}, 0)
//
//	var tx, _ = NewExecuteContractTransaction(
//		privateKey,
//		from,
//		"fe8fc34a2b981fbd86ed11bf27e7d54dfd0fc54a",
//		"",
//		method,
//		params,
//		theTime,
//	)
//	//fmt.Printf("EXECUTE_Get: %s", tx.String())
//}

func TestBadKeyTransaction(t *testing.T) {

	//var tx *Transaction

	//codeBytes := make([]byte, 0)
	//privKeyBytes, err := hex.DecodeString("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25")
	//if err != nil {
	//	t.Log("Correctly determined that this is an invalid key")
	//}

	tx, err := NewTransferTokensTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25",
		"7777f2b40aacbef5a5127f65418dc5f951280833",
		"0e19046b35344383ac0a27c1902fdc1c8c060fa9",
		1,
		0,
		utils.ToMilliSeconds(time.Now()),
		//codeBytes,
	)
	if err != nil {
		t.Log("Correctly failed to create transaction with invalid key")
	}

	if tx != nil {
		t.Error("Created a TX off a bad key.  ")
	}
}

func TestBadFutureTime(t *testing.T) {
	_, err := NewTransferTokensTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
		"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c",
		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
		1,
		0,
		utils.ToMilliSeconds(time.Now()) + int64(10000),
	)

	if err != nil {
		t.Log("Correctly failed to create transaction with bad time value")
		t.Log(err)
	}
}

func TestBadNegativeTime(t *testing.T) {
	_, err := NewTransferTokensTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
		"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c",
		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
		1,
		0,
		-1,
	)

	if err != nil {
		t.Log("Correctly failed to create transaction with bad time value")
		t.Log(err)
	}
}

func TestVerify(t *testing.T) {
	tx := testMockTransaction(t)
	err := tx.Verify()

	if err != nil {
		t.Error("Verify signature is NOT working", err)
	}
}

//TestGettersSetters
func TestGettersSetters(t *testing.T) {
	tx := testMockTransaction(t)
	if tx.Key() == "" {
		t.Error("Key() failed")
	}
	if tx.TypeKey() == "" {
		t.Error("TypeKey() failed")
	}
	if tx.TimeKey() == "" {
		t.Error("TimeKey() failed")
	}
	if tx.FromKey() == "" {
		t.Error("FromKey() failed")
	}

	_, err := tx.MarshalJSON()

	if err != nil {
		t.Error("MarshJSON() failed on transaction")
	}

}

//TestTransactionCalculateHash
func TestTransactionCalculateHash(t *testing.T) {
	// TODO: Transaction.CalculateHash()
	t.Skip("Need a unit test for this...")
}

//TestTransactionEquals
func TestTransactionEquals(t *testing.T) {
	// TODO: Transaction.Equals()
	t.Skip("Need a unit test for this...")
}

//TestTransactionSet
func TestTransactionSet(t *testing.T) {
	// TODO: Transaction.PersistAndCache()
	t.Skip("Need a Badger DB mock")
}

//TestToTransactions
func TestToTransactions(t *testing.T) {
	// TODO: ToTransactions()
	t.Skip("Need a Badger DB mock")
}

//TestToTransactionsByFromAddress
func TestToTransactionsByFromAddress(t *testing.T) {
	// TODO: ToTransactionsByFromAddress()
	t.Skip("Need a Badger DB mock")
}

//TestToTransactionsByToAddress
func TestToTransactionsByToAddress(t *testing.T) {
	// TODO: ToTransactionsByToAddress()
	t.Skip("Need a Badger DB mock")
}

//TestToTransactionsByType
func TestToTransactionsByType(t *testing.T) {
	// TODO: ToTransactionsByType()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactionsByKey(t *testing.T) {utils.ToMilliSeconds(time.Now())
	// TODO: ToTransactionsByKey()
	t.Skip("Need a Badger DB mock")
}

//func TestStatusChange(t *testing.T) {
//
//	 tx := freshMockTransaction(t)
//
//	 if tx.Verify() {
//		 t.Log("transaction verified")
//	 } else {
//		 t.Error("cannot verify transaction")
//	 }
//	 txBytes, err := tx.MarshalJSON()
//
//	 if err != nil {
//		 panic(err)
//	 }
//	 fmt.Printf(string(txBytes))
//	 response, err := http.Post("http://localhost:1975/v1/transactions", "application/json", bytes.NewBuffer(txBytes))
//	 if err != nil {
//		 fmt.Printf("The HTTP request failed with error %s\n", err)
//		 panic(err)
//	 }
//	 data, _ := ioutil.ReadAll(response.Body)
//	 fmt.Printf("Request: \n%v\n", string(data))
//
//	 var x map[string]interface{}
//	 json.Unmarshal(data, &x)
//
//	 response, err = http.Get(fmt.Sprintf("http://localhost:1975/v1/statuses/%v", x["id"]))
//	 if err != nil {
//		 fmt.Printf("The HTTP request failed with error %s\n", err)
//	 } else {
//		 data, _ := ioutil.ReadAll(response.Body)
//		 fmt.Printf("Status: \n%v\n", string(data))
//	 }
// }

//func freshMockTransaction(t *testing.T) *Transaction {
//	tx, err := NewTransferTokensTransaction(
//		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
//		"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c",
//		//"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
//		"",
//		1,
//		0,
//		utils.ToMilliSeconds(time.Now()),
//	)
//	if err != nil {
//		t.Fatalf("Could not create transaction %s", err.Error())
//	}
//	return tx
//}
