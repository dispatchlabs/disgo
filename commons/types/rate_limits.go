package types

import (
	"encoding/json"
	"github.com/dispatchlabs/disgo/commons/utils"
)

// RateLimits - config for our rate limiting system
type RateLimits struct {
	EpochTime   uint64
	NumWindows  uint64
	TxPerMinute uint64
	AvgHzPerTxn uint64
}

var RateLimitsDefaults *RateLimits

func init() {
	RateLimitsDefaults = &RateLimits{
		EpochTime:   1538352000000000000,
		NumWindows:  240,
		TxPerMinute: 600,
		AvgHzPerTxn: 13162215217,
	}
}

// UnmarshalJSON
func (this *RateLimits) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}
	if jsonMap["epochTime"] != nil {
		this.EpochTime = uint64(jsonMap["epochTime"].(float64))
	}
	if jsonMap["numWindows"] != nil {
		this.NumWindows = uint64(jsonMap["numWindows"].(float64))
	}
	if jsonMap["txPerMinute"] != nil {
		this.TxPerMinute = uint64(jsonMap["txPerMinute"].(float64))
	}
	if jsonMap["avgHzPerTxn"] != nil {
		this.AvgHzPerTxn = uint64(jsonMap["avgHzPerTxn"].(float64))
	}

	return nil
}

// MarshalJSON
func (this RateLimits) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EpochTime   uint64 `json:"epochTime"`
		NumWindows  uint64 `json:"numWindows"`
		TxPerMinute uint64 `json:"txPerMinute"`
		AvgHzPerTxn uint64 `json:"avgHzPerTxn"`
	}{
		EpochTime:   this.EpochTime,
		NumWindows:  this.NumWindows,
		TxPerMinute: this.TxPerMinute,
		AvgHzPerTxn: this.AvgHzPerTxn,
	})
}

// String
func (this RateLimits) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal endpoint", err)
		return ""
	}
	return string(bytes)
}
