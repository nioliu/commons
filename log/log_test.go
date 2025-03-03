package log

import (
	"context"
	"go.uber.org/zap"
	"testing"
	"time"
)

// stand for other services
func TestLogWithKafka(t *testing.T) {
	ctx := context.Background()

	// add trace_id for tracing
	ctx = context.WithValue(ctx, "trace_id", "111222333")

	// write log to standard output and kafka
	for i := 0; i < 10000; i++ {
		InfoWithCtxFields(ctx,
			"this is a test message for presentation",
			zap.String("my name", "nioliu"),
			zap.Int("index", i))

	}

	// prevent the main goroutine from exiting
	time.Sleep(1 * time.Minute)
}
