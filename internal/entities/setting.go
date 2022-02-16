package entities

import "encoding/json"

type Setting struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}
