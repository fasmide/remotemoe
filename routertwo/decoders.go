package routertwo

import "encoding/json"

type Decoder interface {
	Decode(json.RawMessage) (interface{}, error)
}

var decoders map[string]Decoder

func RegisterDecoder(n string, d Decoder) {
	if decoders == nil {
		decoders = make(map[string]Decoder)
	}
	decoders[n] = d
}
