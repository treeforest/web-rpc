package jsoncodec

import (
	"encoding/json"
	"github.com/treeforest/web-rpc/codec"
)

type jsonCodec struct {
}

func NewCodec() codec.Codec {
	return &jsonCodec{}
}

func (p *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (p *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
