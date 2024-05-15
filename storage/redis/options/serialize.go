package options

import (
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
)

type Serialize int

const (
	SerializeJSON Serialize = iota
	SerializeMessagePack
)

func (r Serialize) Valid() bool {
	return r >= SerializeJSON && r <= SerializeMessagePack
}

func (r Serialize) Marshal(v any) ([]byte, error) {
	switch r {
	case SerializeMessagePack:
		return msgpack.Marshal(v)
	default:
		return json.Marshal(v)
	}
}

func (r Serialize) Unmarshal(data []byte, v any) error {
	switch r {
	case SerializeMessagePack:
		return msgpack.Unmarshal(data, v)
	default:
		return json.Unmarshal(data, v)
	}
}
