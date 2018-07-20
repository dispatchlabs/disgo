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
	"errors"
	"testing"
	"time"
	"reflect"
)

var testReceiptByte = []byte("{\"id\":\"60ef98ce-73b2-470e-ae92-4cd0a1eae0a3\",\"type\":\"test\",\"status\":\"Pending\",\"humanReadableStatus\":\"Pending\",\"data\":\"test data\",\"contractAddress\":\"\",\"contractResult\":[],\"created\":\"2018-05-09T15:04:05Z\"}")


//TestReceiptCache
func TestReceiptCache(t *testing.T) {
	receipt := NewReceipt("test")
	receipt.Cache(c, time.Second * 5)
	testReceipt, err := ToReceiptFromCache(c, receipt.Id)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(testReceipt, receipt) == false{
		t.Error("Reciept not equal to testReceipt")
	}
}

//TestNewReceipt
func TestNewReceipt(t *testing.T) {
	receipt := NewReceipt("test")
	if receipt.Id == "" {
		t.Errorf("NewReceipt returning invalid %s value: %s", "Id", receipt.Id)
	}
	if receipt.Type != "test" {
		t.Errorf("NewReceipt returning invalid %s value: %s", "Type", receipt.Type)
	}
	if receipt.Status != "Pending" {
		t.Errorf("NewReceipt returning invalid %s value: %s", "Status", receipt.Status)
	}
	if time.Time.IsZero(receipt.Created) {
		t.Error("NewReceipt receipt.Created is empty")
	}

	if receipt.HumanReadableStatus != "" {
		t.Errorf("NewReceipt returning invalid %s value: %s", "HumanReadableStatus", receipt.HumanReadableStatus)
	}
	if receipt.Data != nil {
		t.Errorf("NewReceipt returning invalid %s value: %s", "Data", receipt.Data)
	}
}

//TestNewReceiptWithStatus
func TestNewReceiptWithStatus(t *testing.T) {
	receipt := NewReceiptWithStatus("test", "Pending", "Pending")
	if receipt.Id == "" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Id", receipt.Id)
	}
	if receipt.Type != "test" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Type", receipt.Type)
	}
	if receipt.Status != "Pending" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Status", receipt.Status)
	}
	if receipt.HumanReadableStatus != "Pending" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "HumanReadableStatus", receipt.HumanReadableStatus)
	}
	if time.Time.IsZero(receipt.Created) {
		t.Error("NewReceiptWithStatus receipt.Created is empty")
	}
	if receipt.Data != nil {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Data", receipt.Data)
	}
}

//TestNewReceiptWithError
func TestNewReceiptWithError(t *testing.T) {
	err := errors.New("test error")
	receipt := NewReceiptWithError("test", err)
	if receipt.Id == "" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Id", receipt.Id)
	}
	if receipt.Type != "test" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Type", receipt.Type)
	}
	if receipt.Status != "InternalError" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Status", receipt.Status)
	}
	if receipt.HumanReadableStatus != "test error" {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "HumanReadableStatus", receipt.HumanReadableStatus)
	}
	if time.Time.IsZero(receipt.Created) {
		t.Error("NewReceiptWithStatus receipt.Created is empty")
	}
	if receipt.Data != nil {
		t.Errorf("NewReceiptWithStatus returning invalid %s value: %s", "Data", receipt.Data)
	}
}

//TestToReceiptFromJson
func TestToReceiptFromJson(t *testing.T) {
	receipt, err := ToReceiptFromJson(testReceiptByte)
	if err != nil {
		t.Fatalf("ToReceiptFromJson returning error: %s", err)
	}
	testReceiptStruct(t, receipt)
}

//TestReceiptUnmarshalJSON
func TestReceiptUnmarshalJSON(t *testing.T) {
	receipt := &Receipt{}
	receipt.UnmarshalJSON(testReceiptByte)
	testReceiptStruct(t, receipt)
}

//TestReceiptMarshalJSON
func TestReceiptMarshalJSON(t *testing.T) {
	receipt := &Receipt{}
	receipt.UnmarshalJSON(testReceiptByte)
	out, err := receipt.MarshalJSON()
	if err != nil {
		t.Fatalf("receipt.MarshalJSON returning error: %s", err)
	}
	//if reflect.DeepEqual(out, receipt.String()) == false {
	if string(out) != receipt.String() {
		t.Errorf("receipt.MarshalJSON returning invalid value.\nGot: %s\nExpected: %s", out, receipt.String())
	}
}

//TestToReceiptFromId
func TestToReceiptFromId(t *testing.T) {
	// TODO: ToReceiptFromId()
	t.Skip("Need a Badger DB mock")
}

//TestReceiptSet
func TestReceiptSet(t *testing.T) {
	// TODO: Receipt.PersistAndCache()
	t.Skip("Need a Badger DB mock")
}

//TestReceiptSetInternalErrorWithNewTransaction
func TestReceiptSetInternalErrorWithNewTransaction(t *testing.T) {
	// TODO: receipt.SetInternalErrorWithNewTransaction()
	t.Skip("Need a Badger DB mock")
}

//TestReceiptSetStatusWithNewTransaction
func TestReceiptSetStatusWithNewTransaction(t *testing.T) {
	// TODO: receipt.SetStatusWithNewTransaction()
	t.Skip("Need a Badger DB mock")
}

//testReceiptStruct
func testReceiptStruct(t *testing.T, receipt *Receipt) {
	if receipt.Id != "60ef98ce-73b2-470e-ae92-4cd0a1eae0a3" {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "Id", receipt.Id)
	}
	if receipt.Type != "test" {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "Type", receipt.Type)
	}
	if receipt.Status != "Pending" {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "Status", receipt.Status)
	}
	if receipt.HumanReadableStatus != "Pending" {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "HumanReadableStatus", receipt.HumanReadableStatus)
	}
	if receipt.Data != "test data" {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "Data", receipt.Data)
	}
	d, _ := time.Parse(time.RFC3339, "2018-05-09T15:04:05Z")
	if receipt.Created != d {
		t.Errorf("receipt.UnmarshalJSON returning invalid %s value: %s", "Created", receipt.Created.String())
	}
	if receipt.Key() != "table-receipt-60ef98ce-73b2-470e-ae92-4cd0a1eae0a3" {
		t.Errorf("receipt.Key() returning invalid %s value: %s", "Key", receipt.Key())
	}
}
