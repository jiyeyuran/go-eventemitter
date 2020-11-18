package eventemitter

import "encoding/json"

// Decoder defines interface to decode bytes into golang type
type Decoder interface {
	Decode(data []byte, result interface{}) error
}

// JsonDecoder implements Decoder with standard json decoder.
type JsonDecoder struct{}

func (JsonDecoder) Decode(data []byte, result interface{}) error {
	return json.Unmarshal(data, result)
}
