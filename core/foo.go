package core

import (
	"encoding/json"
)

type Foo struct {
	Name string
}

func (foo *Foo) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, foo)
}

func (foo Foo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{
		Name string `json:"name"`
	}{
		Name: foo.Name,
	})
}