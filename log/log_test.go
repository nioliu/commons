package log

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	InfoWithCtxFields(context.Background(), "这是信息")
}
