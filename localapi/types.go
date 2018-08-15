package localapi

// Transfer -
type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

// Deploy -
type Deploy struct {
	ByteCode string `json:"byteCode"`
	Abi      string `json:"abi"`
}

// Execute -
type Execute struct {
	ContractAddress string        `json:"contractAddress"`
	Abi             string        `json:"abi"`
	Method          string        `json:"method"`
	Params          []interface{} `json:"params"`
}
