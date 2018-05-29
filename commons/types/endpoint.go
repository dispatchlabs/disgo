package types

import (
	"encoding/json"
	"github.com/dispatchlabs/commons/utils"
)

// Endpoint - Is the DisGover's notion of where a node can be contacted
type Endpoint struct {
	Host string
	Port int64
}

// UnmarshalJSON
func (this *Endpoint) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}
	if jsonMap["host"] != nil {
		this.Host = jsonMap["host"].(string)
	}
	if jsonMap["port"] != nil {
		this.Port = int64(jsonMap["port"].(float64))
	}
	return nil
}

// MarshalJSON
func (this Endpoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Host string `json:"host"`
		Port int64  `json:"port"`
	}{
		Host: this.Host,
		Port: this.Port,
	})
}

// String
func (this Endpoint) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal endpoint", err)
		return ""
	}
	return string(bytes)
}
