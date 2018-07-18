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
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

// Name
type Receipt struct {
	Id                  string
	Type                string
	Status              string
	HumanReadableStatus string
	Data                interface{}
	ContractAddress     string
	ContractResult      []interface{}
	Created             time.Time
}


// Key
func (this Receipt) Key() string {
	return fmt.Sprintf("table-receipt-%s", this.Id)
}

//Cache
func (this *Receipt) Cache(cache *cache.Cache,time_optional ...time.Duration){
	TTL := ReceiptTTL
	if len(time_optional) > 0 {
		TTL = time_optional[0]
	}
	cache.Set(this.Key(), this, TTL)
}

//Persist
func (this *Receipt) Persist(txn *badger.Txn) error{
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	return nil
}

// Set
func (this *Receipt) Set(txn *badger.Txn,cache *cache.Cache) error {
	this.Cache(cache)

	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

// Unset
func (this *Receipt) Unset(txn *badger.Txn,cache *cache.Cache) error {
	cache.Delete(this.Key())
	err := txn.Delete([]byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON
func (this *Receipt) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["id"] != nil {
		this.Id = jsonMap["id"].(string)
	}
	if jsonMap["type"] != nil {
		this.Type = jsonMap["type"].(string)
	}
	if jsonMap["status"] != nil {
		this.Status = jsonMap["status"].(string)
	}
	if jsonMap["humanReadableStatus"] != nil {
		this.HumanReadableStatus = jsonMap["humanReadableStatus"].(string)
	}
	if jsonMap["data"] != nil {
		this.Data = jsonMap["data"]
	}
	if jsonMap["contractAddress"] != nil && jsonMap["contractAddress"] != "" {
		this.ContractAddress = jsonMap["contractAddress"].(string)
	}
	if jsonMap["contractResult"] != nil {
		var contractResult = jsonMap["contractResult"]
		this.ContractResult = contractResult.([]interface{})
	}
	if jsonMap["created"] != nil {
		created, err := time.Parse(time.RFC3339, jsonMap["created"].(string))
		if err != nil {
			return err
		}
		this.Created = created
	}
	return nil
}

// MarshalJSON
func (this Receipt) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id                  string        `json:"id"`
		Type                string        `json:"type"`
		Status              string        `json:"status"`
		HumanReadableStatus string        `json:"humanReadableStatus,omitempty"`
		Data                interface{}   `json:"data,omitempty"`
		ContractAddress     string        `json:"contractAddress,omitempty"`
		ContractResult      []interface{} `json:"contractResult,omitempty"`
		Created             time.Time     `json:"created"`
	}{
		Id:                  this.Id,
		Type:                this.Type,
		Status:              this.Status,
		HumanReadableStatus: this.HumanReadableStatus,
		Data:                this.Data,
		ContractAddress:     this.ContractAddress,
		ContractResult:      this.ContractResult,
		Created:             this.Created,
	})
}

// String
func (this Receipt) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal receipt", err)
		return ""
	}
	return string(bytes)
}

// SetInternalErrorWithNewTransaction
func (this *Receipt) SetInternalErrorWithNewTransaction(db *badger.DB, err error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	this.Status = StatusInternalError
	this.HumanReadableStatus = err.Error()
	err = txn.SetWithTTL([]byte(this.Key()), []byte(this.String()), ReceiptTTL)
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
}

// SetStatusWithNewTransaction
func (this *Receipt) SetStatusWithNewTransaction(db *badger.DB, status string) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	this.Status = status
	err := txn.SetWithTTL([]byte(this.Key()), []byte(this.String()), ReceiptTTL)
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
}


// NewReceipt
func NewReceipt(tipe string) *Receipt {
	return &Receipt{Id: uuid.New().String(), Type: tipe, Status: StatusPending, Created: time.Now()}
}

// NewReceiptWithStatus
func NewReceiptWithStatus(tipe string, status string, humanReadableStatus string) *Receipt {
	return &Receipt{Id: uuid.New().String(), Type: tipe, Status: status, HumanReadableStatus: humanReadableStatus, Created: time.Now()}
}

// NewReceiptWithError
func NewReceiptWithError(tipe string, err error) *Receipt {
	return &Receipt{Id: uuid.New().String(), Type: tipe, Status: StatusInternalError, HumanReadableStatus: err.Error(), Created: time.Now()}
}

// ToReceiptFromJson
func ToReceiptFromJson(payload []byte) (*Receipt, error) {
	receipt := &Receipt{}
	err := json.Unmarshal(payload, receipt)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

// ToReceiptFromCache -
func ToReceiptFromCache(cache *cache.Cache, id string) (*Receipt, error) {
	value, ok :=cache.Get(fmt.Sprintf("table-receipt-%s", id))
	if !ok{
		return nil, ErrNotFound
	}
	receipt := value.(*Receipt)
	return receipt, nil
}

// ToReceiptFromId
func ToReceiptFromId(txn *badger.Txn, id string) (*Receipt, error) {
	item, err := txn.Get([]byte("table-receipt-" + id))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	receipt, err := ToReceiptFromJson(value)
	if err != nil {
		return nil, err
	}
	return receipt, err
}
