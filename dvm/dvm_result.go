package dvm

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"github.com/dispatchlabs/disgo/dvm/ethereum/types"
	"github.com/dispatchlabs/disgo/dvm/ethereum/vm"
	"github.com/dispatchlabs/disgo/dvm/vmstatehelperimplemtations"
)

// DVMResult - represents a result after a `Deploy` or `Execute` of a Smart Contract
type DVMResult struct {
	From                     crypto.AddressBytes                       //
	To                       crypto.AddressBytes                       //
	ABI                      string                                    // The ABI for the smart contract
	StorageState             *vmstatehelperimplemtations.VMStateHelper // The state of the storage
	ContractAddress          crypto.AddressBytes
	ContractMethod           string // Method
	ContractMethodExecError  error  // Method Execution error
	ContractMethodExecResult []byte // Method Execution result - parsable with `jsonABI.Unpack`

	Divvy               int64
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
// func (this *DVMResult) UnmarshalJSON(bytes []byte) error {
// 	var jsonMap map[string]interface{}
// 	error := json.Unmarshal(bytes, &jsonMap)
// 	if error != nil {
// 		return error
// 	}
// 	if jsonMap["from"] != nil {
// 		this.From = crypto.GetAddressBytes(jsonMap["from"].(string))
// 	}
// 	if jsonMap["contractAddress"] != nil {
// 		this.To = crypto.GetAddressBytes(jsonMap["to"].(string))
// 	}
// 	if jsonMap["hertzCost"] != nil {
// 		this.HertzCost = jsonMap["hertzCost"].(uint64)
// 	}

// 	return nil
// }

// MarshalJSON -
func (this DVMResult) MarshalJSON() ([]byte, error) {

	var methodResult = ""

	if this.ContractMethodExecError == nil {
		// Try read the execution result
		if len(strings.TrimSpace(this.ABI)) > 0 {
			fromHexAsByteArray, _ := hex.DecodeString(this.ABI)
			abiAsString := string(fromHexAsByteArray)
			jsonABI, err := abi.JSON(strings.NewReader(abiAsString))
			if err == nil {
				var parsedRes string
				err = jsonABI.Unpack(&parsedRes, this.ContractMethod, this.ContractMethodExecResult)
				if err == nil {
					methodResult = parsedRes
				}
			}
		}
	}

	buf := new(bytes.Buffer)
	vm.WriteLogs(buf, this.Logs)

	return json.Marshal(struct {
		From                     string `json:"from"`
		To                       string `json:"to"`
		ContractAddress          string `json:"contractAddress"`
		ContractMethod           string `json:"ContractMethod"`
		ContractMethodExecResult string `json:"contractMethodExecResult"`
		Divvy                    int64  `json:"divvy"`
		HertzCost                uint64 `json:"hertzCost"`
		Logs                     string `json:"logs"`
	}{
		From:                     crypto.Encode(this.From[:]),
		To:                       crypto.Encode(this.To[:]),
		ContractAddress:          crypto.Encode(this.ContractAddress[:]),
		ContractMethod:           this.ContractMethod,
		ContractMethodExecResult: methodResult,
		Divvy:                    this.Divvy,
		HertzCost:                this.HertzCost,
		Logs:                     buf.String(),
	})
}
