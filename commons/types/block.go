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
)

// Block
type Block struct {
	Id                   int64
	Hash                 string
	NumberOfTransactions int64
	Updated              time.Time
	Created              time.Time
}

// Key
func (this Block) Key() string {
	return fmt.Sprintf("block-%d", this.Id)
}

// UnmarshalJSON
func (this *Block) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["id"] != nil {
		this.Id = int64(jsonMap["id"].(float64))
	}
	if jsonMap["hash"] != nil {
		this.Hash = jsonMap["hash"].(string)
	}
	if jsonMap["numberOfTransactions"] != nil {
		this.NumberOfTransactions = int64(jsonMap["numberOfTransactions"].(float64))
	}
	if jsonMap["updated"] != nil {
		updated, err := time.Parse(time.RFC3339, jsonMap["updated"].(string))
		if err != nil {
			return err
		}
		this.Updated = updated
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
func (this Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id                   int64  `json:"id"`
		Hash                 string `json:"hash"`
		NumberOfTransactions int64  `json:"numberOfTransactions"`
		Updated              string `json:"updated"`
		Created              string `json:"created"`
	}{
		Id:                   this.Id,
		Hash:                 this.Hash,
		NumberOfTransactions: this.NumberOfTransactions,
		Updated:              this.Updated.Format(time.RFC3339),
		Created:              this.Created.Format(time.RFC3339),
	})
}
