package codec

import "reflect"

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
