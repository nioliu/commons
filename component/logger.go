package component

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func GetStandardLogger(timeZone, timeFormat, timeKey, stackTraceKey string) (*zap.Logger, error) {

	// set some option
	config := zap.NewProductionEncoderConfig()

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
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

	return productionConfig.Build()
}
