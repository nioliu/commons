package codec

import (
	"encoding/json"
	"reflect"
)

type JsonCodec struct {
}

func (j *JsonCodec) Name() string {
	return j.String()
}

func (j *JsonCodec) Marshal(v interface{}) ([]byte, error) {
	t := reflect.ValueOf(v)
	if t.Kind() == reflect.Pointer {
		if t.IsNil() {
			return nil, nil
		}
		t = t.Elem()
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
	t := reflect.ValueOf(v)
	if t.Kind() != reflect.Pointer {
		return NotPointer
	}
	t = t.Elem()
	switch t.Kind() {
	case reflect.Slice:
		t.SetBytes(data)
		return nil
	case reflect.Struct:
		return json.Unmarshal(data, v)
	default:
		return UnSupportedType
	}
}

func (j *JsonCodec) String() string {
	return "json"
}
