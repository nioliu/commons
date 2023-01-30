package component

import (
	"strings"
	"testing"
)

func TestFromByteToStr(t *testing.T) {
	b := "随意字符串"
	t.Log([]byte(b))
	fromByteToStr, err := FromByteToStr("[233 154 143 230 132 143 229 173 151 231 172 166 228 184 178]")
	if err != nil {
		panic(err)
	}
	t.Log(string(fromByteToStr))
}
func TestDWQ(t *testing.T) {
	mark := strings.SplitN("byte:{21312}", ",", 2)
	t.Log(len(mark))
}
