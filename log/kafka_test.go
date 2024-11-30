package log

import (
	"go.uber.org/zap/zapcore"
	"reflect"
	"testing"
)

func Test_withKafkaCore(t *testing.T) {
	type args struct {
		ec *zapcore.EncoderConfig
	}
	tests := []struct {
		name string
		args args
		want *kafkaCore
	}{
		{
			name: "Test_withKafkaCore",
			args: args{
				ec: &zapcore.EncoderConfig{},
			},
			want: &kafkaCore{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withKafkaCore(tt.args.ec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("withKafkaCore() = %v, want %v", got, tt.want)
			}
		})
	}
}
