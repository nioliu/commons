package grpc

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	UnMatchedCodecType = errors.New("unmatched codec type for current req and rsp type")

	NotPointer = errors.New("parameter is not pointer")
)

type JsonCodec struct {
}

func (j *JsonCodec) Marshal(v interface{}) ([]byte, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func (j *JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *JsonCodec) String() string {
	return "json"
}

type ByteCodec struct {
}

func (b *ByteCodec) Marshal(v interface{}) ([]byte, error) {
	bytes, ok := v.([]byte)
	if !ok {
		return nil, UnMatchedCodecType
	}
	return bytes, nil
}

func (b *ByteCodec) Unmarshal(data []byte, v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Pointer {
		return NotPointer
	}
	i := v.(*[]byte)
	*i = data
	return nil
}

func (b *ByteCodec) String() string {
	return "byte"
}
