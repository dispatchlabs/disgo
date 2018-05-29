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
)

func testMockTransaction(t *testing.T) *Transaction {
	//codeBytes := make([]byte, 0)
	d, _ := time.Parse(time.RFC3339, "2018-05-09T15:04:05Z")
	//privKeyBytes, err := hex.DecodeString("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a")
	//if err != nil {
	//	t.Fatalf("Could not create privKeyBytes %s", err.Error())
	//	return nil
	//} else {
		tx, err := NewTransaction(
			"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
			0,
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
	hash := tx.NewHash()

	if hash == "" {
		t.Error("unable to create new hash for a TX")
	}

	if len(hash) != 64 {
		t.Error("hash length is NOT valid")
	}
}

func TestBadKeyTransaction(t *testing.T) {

	//var tx *Transaction

	//codeBytes := make([]byte, 0)
	//privKeyBytes, err := hex.DecodeString("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25")
	//if err != nil {
	//	t.Log("Correctly determined that this is an invalid key")
	//}

	tx, err := NewTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25",
		TransactionTypeTransferTokens,
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
