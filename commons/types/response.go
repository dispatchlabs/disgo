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
	"time"

	"github.com/dispatchlabs/disgo/commons/utils"
)

// Name
type Response struct {
	Status              string
	HumanReadableStatus string
	Data                interface{}
	Created             time.Time
}

// UnmarshalJSON
func (this *Response) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
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
func (this Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Status              string      `json:"status"`
		HumanReadableStatus string      `json:"humanReadableStatus,omitempty"`
		Data                interface{} `json:"data,omitempty"`
		Created             time.Time   `json:"created"`
	}{
		Status:              this.Status,
		HumanReadableStatus: this.HumanReadableStatus,
		Data:                this.Data,
		Created:             this.Created,
	})
}

// String
func (this Response) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal receipt", err)
		return ""
	}
	return string(bytes)
}

// NewResponse
func NewResponse() *Response {
	return &Response{Status: StatusOk, HumanReadableStatus: "Ok", Created: time.Now()}
}

// NewResponseWithError
func NewResponseWithError(err error) *Response {
	return &Response{Status: StatusInternalError, HumanReadableStatus: err.Error(), Created: time.Now()}
}

// NewResponseWithStatus
func NewResponseWithStatus(status string, humanReadableStatus string) *Response {
	return &Response{Status: status, HumanReadableStatus: humanReadableStatus, Created: time.Now()}
}
