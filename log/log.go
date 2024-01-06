package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"time"
)

type LoggerConfig struct {
	engine            *zap.Logger            // 实际的logger引擎
	ctxFields         map[string]interface{} // fields need to be printed from context, map[name]ctxIndex
	printGrpcMetadata bool                   // choose if print grpc metadata, it also should set the key and value to ctxFields
}

const WithKafka = "WITH_KAFKA"

var logger = &LoggerConfig{}
var err error

func init() {
	logger = newDefaultLogger()
}

func newDefaultLogger() *LoggerConfig {
	l := &LoggerConfig{}
	l.engine, err = getEngine("Asia/Shanghai",
		"2006-01-02 15:04:05 Z07", "time", "")
	if err != nil {
		log.Fatal("init logger failed", zap.Error(err))
	}
	l.engine = l.engine.WithOptions(zap.AddCallerSkip(1))

	//l.WithGrpcMetadata()
	l.WithContextFields(getStandardCtxFieldsMap())

	return l
}

func GetRawLogger() *zap.Logger {
	return logger.engine
}

func (logger *LoggerConfig) WithContextFields(ctxFields map[string]interface{}) {
	logger.ctxFields = ctxFields
}

func (logger *LoggerConfig) WithGrpcMetadata() {
	logger.printGrpcMetadata = true
}

func (logger *LoggerConfig) WithLoggerEngine(engine *zap.Logger) {
	logger.engine = engine
}

func getStandardCtxFieldsMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["trace_id"] = "trace_id"

	return m
}

// getEngine 包含标准输出以及根据环境变量是否使用kafka
func getEngine(timeZone, timeFormat, timeKey, stackTraceKey string) (*zap.Logger, error) {
	cores := make([]zapcore.Core, 0)
	standardCore, encoderConfig, err := withStandardCore(timeZone, timeFormat, timeKey, stackTraceKey)
	if err != nil {
		return nil, err
	} else {
		cores = append(cores, standardCore.Core())
	}
	// with kafka
	if os.Getenv(WithKafka) == "true" {
		cores = append(cores, withKafkaCore(encoderConfig))
	}
	return zap.New(zapcore.NewTee(cores...)), nil
}

func withStandardCore(timeZone string, timeFormat string, timeKey string,
	stackTraceKey string) (*zap.Logger, *zapcore.EncoderConfig, error) {
	// set some option
	config := zap.NewProductionEncoderConfig()

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, nil, err
	}

	config.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		t = t.In(location)
		t.Format(timeFormat)
		encoder.AppendString(t.String())
	}

	config.TimeKey = timeKey
	config.StacktraceKey = stackTraceKey
	productionConfig := zap.NewProductionConfig()
	productionConfig.EncoderConfig = config
	build, err := productionConfig.Build()
	if err != nil {
		return nil, nil, err
	}

	return build, &config, nil
}

// extract key and value from config, and set each field's name to key, value to map's value
func (logger *LoggerConfig) getKvFromCtx(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, len(logger.ctxFields))

	var inCtx = metadata.MD{}
	existIn := false
	var outCtx = metadata.MD{}
	existOut := false
	if logger.printGrpcMetadata {
		inCtx, existIn = metadata.FromIncomingContext(ctx)
		outCtx, existOut = metadata.FromOutgoingContext(ctx)
	}

	for k, v := range logger.ctxFields {
		fieldValue := ctx.Value(v)

		// extract from grpc metadata
		if logger.printGrpcMetadata {
			if vs, ok := v.(string); ok {
				if existIn && inCtx.Get(vs) != nil {
					fieldValue = inCtx.Get(vs)
				}
				if existOut && outCtx.Get(vs) != nil {
					fieldValue = outCtx.Get(vs)
				}
			}
		}

		if fieldValue != "" {
			fields = append(fields, zap.Any(k, fieldValue))
		}
	}

	return fields
}

func getDefaultLogger() *LoggerConfig {
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

func DebugWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.Debug(msg, fields...)
}

func InfoWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.Info(msg, fields...)
}

func WarnWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.Warn(msg, fields...)
}

func ErrorWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.Error(msg, fields...)
}

func DPanicWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.DPanic(msg, fields...)
}

func FatalWithCtxFields(ctx context.Context, msg string, fields ...zap.Field) {
	logger := getDefaultLogger()
	if ctx != nil {
		fields = append(fields, logger.getKvFromCtx(ctx)...)
	}
	logger.engine.Fatal(msg, fields...)
}
