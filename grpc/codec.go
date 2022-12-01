package grpc

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	UnMatchedCodecType = errors.New("unmatched codec type for current req and rsp type")
	UnSupportedType    = errors.New("unsupported input or output type")
	NotPointer         = errors.New("parameter is not pointer")
)

type JsonCodec struct {
}

func (j *JsonCodec) Name() string {
	return j.String()
}

func (j *JsonCodec) Marshal(v interface{}) ([]byte, error) {
	t := reflect.ValueOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.IsNil() {
		return nil, nil
	}
	switch t.Kind() {
	case reflect.Slice:
		return t.Bytes(), nil
	case reflect.Struct:
		marshal, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return marshal, nil
	default:
		return nil, UnSupportedType
	}
}

func (j *JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *JsonCodec) String() string {
	return "json"
}

type ByteCodec struct {
}

func (b *ByteCodec) Name() string {
	return b.String()
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
