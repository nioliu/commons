package component

import (
	"errors"
	"strconv"
	"strings"
)

// FromByteToStandard From log byte to standard []byte, the format of byte is "[01 23 31 21 ...]"
func FromByteToStandard(b string) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	if b[0] == '[' && b[len(b)-1] == ']' {
		b = b[1 : len(b)-1]
	}
	split := strings.Split(b, " ")
	res := make([]byte, 0, len(split))
	for _, v := range split {
		atoi, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.New("illegal b format")
		}
		res = append(res, byte(atoi))
	}
	return res, nil
}
