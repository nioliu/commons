package log

import (
	"context"
	"encoding/json"
	"log"
	"testing"
)

func TestLog(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "123")
	marshal, err2 := json.Marshal(ctx.Value(""))
	if err2 != nil {
		log.Fatalln(err2)
	}
	t.Log(string(marshal))
	InfoWithCtxFields(ctx, "这是信息")
}

func TestTopic(t *testing.T) {
	for i := 0; i < 100; i++ {
		go func() { println(getTopic()) }()
	}
}
