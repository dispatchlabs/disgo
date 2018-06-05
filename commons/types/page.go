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
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"strconv"
	"github.com/patrickmn/go-cache"
)

// Block
type Page struct {
	Number      		int64
	TransactionsHash    string
	StateHash			string
	ParentHash 			string
	BWused				int64
}

//TODO: link transactions to pages... maybe transaction Page key.

// Key
func (this Page) Key() string {
	return fmt.Sprintf("Page-%d", this.Number)
}

//Cache
func (this *Page) Cache(cache *cache.Cache){
	cache.Set(strconv.Itoa(int(this.Number)), this, PageTTL)
}

//Persist
func (this *Page) Persist(txn *badger.Txn) error{
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	return nil
}

// Set
func (this *Page) Set(txn *badger.Txn,cache *cache.Cache) error {
	this.Cache(cache)

	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

func (this *Page) Delete(txn *badger.Txn) error {
	err := txn.Delete([]byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}


// UnmarshalJSON
func (this *Page) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["number"] != nil {
		this.Number = int64(jsonMap["number"].(float64))
	}
	if jsonMap["transactionsHash"] != nil {
		this.TransactionsHash = jsonMap["transactionsHash"].(string)
	}
	if jsonMap["stateHash"] != nil {
		this.StateHash = jsonMap["stateHash"].(string)
	}
	if jsonMap["parentHash"] != nil {
		this.ParentHash = jsonMap["parentHash"].(string)
	}
	if jsonMap["bwUsed"] != nil {
		this.BWused = int64(jsonMap["bwUsed"].(float64))
	}

	return nil
}

// MarshalJSON
func (this Page) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Number               int64  `json:"number"`
		TransactionsHash     string `json:"transactionsHash"`
		StateHash 			 string  `json:"stateHash"`
		ParentHash           string `json:"parentHash"`
		BWused               int64 `json:"bwUsed"`
	}{
		Number:                   this.Number,
		TransactionsHash:                 this.TransactionsHash,
		StateHash: this.StateHash,
		ParentHash:              this.ParentHash,
		BWused:              this.BWused,
	})
}

// String
func (this Page) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal node", err)
		return ""
	}
	return string(bytes)
}



func ToPageFromJson(payload []byte) (*Page, error) {
	page := &Page{}
	err := json.Unmarshal(payload, page)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func ToPageByKey(txn *badger.Txn, key []byte) (*Page, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	node, err := ToPageFromJson(value)
	if err != nil {
		return nil, err
	}
	return node, err
}