package log

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// kafkaCore is a custom zap core for sending logs to Kafka.
type kafkaCore struct {
	encoder     zapcore.Encoder
	writer      *kafka.Writer
	level       zapcore.LevelEnabler
	serviceName string
}

const batchSize = 20
const globalKeyEnv = "SERVICE_NAME"

// 默认的日志topics
var defaultLogTopics = []string{"log-1", "log-2", "log-3"}

var currTopicIndex = atomic.Int32{}

var brokers = []string{"b2s-kafka:9092"}

func withKafkaCore(ec *zapcore.EncoderConfig) *kafkaCore {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{}, // 选择数据量最小的分区写入
		BatchSize:              batchSize,           // 批量发送设置
		AllowAutoTopicCreation: true,                // 自动创建topic
		Async:                  true,
	}

	return &kafkaCore{
		encoder:     zapcore.NewJSONEncoder(*ec),
		writer:      writer,
		level:       zap.InfoLevel,
		serviceName: os.Getenv(globalKeyEnv),
	}
}

// 轮询topic，使得负载均衡，当新增一个topic时，就将之前的topic全删掉，然后等到量差不多了，再使用全topics
func getTopic() string {
	nextTopic := (currTopicIndex.Load() + 1) % int32(len(defaultLogTopics))
	currTopicIndex.Store(nextTopic)
	return defaultLogTopics[nextTopic]
}

func (c *kafkaCore) Enabled(lvl zapcore.Level) bool {
	return c.level.Enabled(lvl)
}

func (c *kafkaCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	for _, f := range fields {
		clone.encoder.AddReflected(f.Key, f.Interface)
	}
	return clone
}

func (c *kafkaCore) clone() *kafkaCore {
	return &kafkaCore{
		encoder: c.encoder.Clone(),
		writer:  c.writer,
		level:   c.level,
	}
}

func (c *kafkaCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

const retryHeaderKey = "retry_times"

func (c *kafkaCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	go func() {
		if err := c.writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(c.serviceName),
			Topic: getTopic(),
			Value: buf.Bytes(),
			Time:  time.Now(),
			// 加入重试标识
			Headers: []kafka.Header{{
				Key:   retryHeaderKey,
				Value: []byte(strconv.Itoa(0)),
			}},
		}); err != nil {
			io.WriteString(os.Stdout, fmt.Sprintf("Log Error: %s", err.Error()))
		}
	}()
	return nil
}

func (c *kafkaCore) Sync() error {
	return c.writer.Close()
}
