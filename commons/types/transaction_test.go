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
	"testing"
	"time"
	"github.com/dispatchlabs/disgo/commons/utils"
	"fmt"
)

func testMockTransaction(t *testing.T) *Transaction {
	//codeBytes := make([]byte, 0)
	d, _ := time.Parse(time.RFC3339, "2018-05-09T15:04:05Z")
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
		d.UnixNano(),
	)

	if err != nil {
		t.Fatalf("Could not create transaction %s", err.Error())
	}
	return tx
	//}
}

// TestTransactionVerify
func TestTransactionVerify(t *testing.T) {
	tx := testMockTransaction(t)

	if tx.Verify() {
		t.Log("transaction verified")
	} else {
		t.Error("cannot verify transaction")
	}
}

func TestNewTransaction(t *testing.T) {
	tx := testMockTransaction(t)
	if tx == nil {
		t.Error("Unable to create Transaction")
	}
	if tx.Signature == "" {
		t.Error("Unable to create Signature on Transaction")
	}
}

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

func TestPrintTransaction(t *testing.T) {
	var privateKey= "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var from= "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var theTime= utils.ToMilliSeconds(time.Now())
	//var method = "getVar5"
	//var params = make([]interface{}, 0)

	var tx, _ = NewTransferTokensTransaction(
		privateKey,
		from,
		"",
		1,
		1,
		theTime,
	)
	fmt.Printf("EXECUTE_Get: %s", tx.String())
}

func TestPrintNewDeployTx(t *testing.T) {
	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var code = "608060405234801561001057600080fd5b506040805190810160405280600d81526020017f61616161616161616161616161000000000000000000000000000000000000008152506000908051906020019061005c9291906100f7565b50600060016000018190555060006001800160006101000a81548160ff02191690831515021790555060018060010160016101000a81548160ff021916908360ff1602179055506040805190810160405280600b81526020017f6262626262626262626262000000000000000000000000000000000000000000815250600160020190805190602001906100f19291906100f7565b5061019c565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013857805160ff1916838001178555610166565b82800160010185558215610166579182015b8281111561016557825182559160200191906001019061014a565b5b5090506101739190610177565b5090565b61019991905b8082111561019557600081600090555060010161017d565b5090565b90565b610664806101ab6000396000f300608060405260043610610078576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806333e538e91461007d57806334e45f531461010d5780633a458b1f1461017657806378d8866e1461022557806379af6473146102b5578063cb69e300146102cc575b600080fd5b34801561008957600080fd5b50610092610335565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100d25780820151818401526020810190506100b7565b50505050905090810190601f1680156100ff5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561011957600080fd5b50610174600480360381019080803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506103d7565b005b34801561018257600080fd5b5061018b6103f4565b60405180858152602001841515151581526020018360ff1660ff16815260200180602001828103825283818151815260200191508051906020019080838360005b838110156101e75780820151818401526020810190506101cc565b50505050905090810190601f1680156102145780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b34801561023157600080fd5b5061023a6104c4565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561027a57808201518184015260208101905061025f565b50505050905090810190601f1680156102a75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156102c157600080fd5b506102ca610562565b005b3480156102d857600080fd5b50610333600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290505050610579565b005b606060008054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156103cd5780601f106103a2576101008083540402835291602001916103cd565b820191906000526020600020905b8154815290600101906020018083116103b057829003601f168201915b5050505050905090565b80600160020190805190602001906103f0929190610593565b5050565b60018060000154908060010160009054906101000a900460ff16908060010160019054906101000a900460ff1690806002018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156104ba5780601f1061048f576101008083540402835291602001916104ba565b820191906000526020600020905b81548152906001019060200180831161049d57829003601f168201915b5050505050905084565b60008054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561055a5780601f1061052f5761010080835404028352916020019161055a565b820191906000526020600020905b81548152906001019060200180831161053d57829003601f168201915b505050505081565b600160000160008154809291906001019190505550565b806000908051906020019061058f929190610593565b5050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106105d457805160ff1916838001178555610602565b82800160010185558215610602579182015b828111156106015782518255916020019190600101906105e6565b5b50905061060f9190610613565b5090565b61063591905b80821115610631576000816000905550600101610619565b5090565b905600a165627a7a72305820f782ba3879cbbd0ec37bd4bbfbe885796488e7504e9b2f1f6817a4d3b8"
	var theTime = utils.ToMilliSeconds(time.Now())

	var tx, _ = NewDeployContractTransaction(
		privateKey,
		from,
		code,
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
		time.Now().UnixNano(),
		//codeBytes,
	)
	if err != nil {
		t.Log("Correctly failed to create transaction with invalid key")
	}

	if tx != nil {
		t.Error("Created a TX off a bad key.  ")
	}
}

func TestBadTime(t *testing.T) {
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
	b := tx.Verify()

	if !b {
		t.Error("Verify signature is NOT working")
	}
}

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

func TestTransactionCalculateHash(t *testing.T) {
	// TODO: Transaction.CalculateHash()
	t.Skip("Need a unit test for this...")
}

func TestTransactionEquals(t *testing.T) {
	// TODO: Transaction.Equals()
	t.Skip("Need a unit test for this...")
}

func TestTransactionSet(t *testing.T) {
	// TODO: Transaction.Set()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactions(t *testing.T) {
	// TODO: ToTransactions()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactionsByFromAddress(t *testing.T) {
	// TODO: ToTransactionsByFromAddress()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactionsByToAddress(t *testing.T) {
	// TODO: ToTransactionsByToAddress()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactionsByType(t *testing.T) {
	// TODO: ToTransactionsByType()
	t.Skip("Need a Badger DB mock")
}

func TestToTransactionsByKey(t *testing.T) {
	// TODO: ToTransactionsByKey()
	t.Skip("Need a Badger DB mock")
}
