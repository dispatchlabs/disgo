package dvm

import (
	"encoding/json"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/types"
)

type DVMResult struct {
	From  crypto.AddressBytes
	To    crypto.AddressBytes
	Divvy int64

	ContractAddress     crypto.AddressBytes
	Status              uint
	HertzCost           uint64
	CumulativeHertzUsed uint64
	Bloom               types.Bloom
	Logs                []*types.Log
}

// String -
func (this DVMResult) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal transaction", err)
		return ""
	}
	return string(bytes)
}

// UnmarshalJSON -
func (this *DVMResult) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["from"] != nil {
		this.From = crypto.GetAddressBytes(jsonMap["from"].(string))
	}
	if jsonMap["contractAddress"] != nil {
		this.ContractAddress = crypto.GetAddressBytes(jsonMap["contractAddress"].(string))
	}
	if jsonMap["hertzCost"] != nil {
		this.HertzCost = jsonMap["hertzCost"].(uint64)
	}

	return nil
}

// MarshalJSON -
func (this DVMResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		From            string `json:"from"`
		ContractAddress string `json:"contractAddress"`
		HertzCost       uint64 `json:"hertzCost"`
	}{
		From:            crypto.Encode(this.From[:]),
		ContractAddress: crypto.Encode(this.ContractAddress[:]),
		HertzCost:       this.HertzCost,
	})
}
