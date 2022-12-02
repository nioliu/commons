package codec

import (
	"errors"
)

var (
	UnMatchedCodecType = errors.New("unmatched codec type for current req and rsp type")
	UnSupportedType    = errors.New("unsupported input or output type")
	NotPointer         = errors.New("parameter is not pointer")
)
