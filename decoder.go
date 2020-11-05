package eventemitter

import "encoding/json"

// Decoder define interface to decode bytes to golang type
type Decoder interface {
	Decode(data []byte, result interface{}) error
}

// JsonDecoder decode json bytes to golang type
type JsonDecoder struct{}

func (JsonDecoder) Decode(data []byte, result interface{}) error {
	return json.Unmarshal(data, result)
}
