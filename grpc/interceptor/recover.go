package interceptor

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"runtime"
)

func GetRecoverInterceptor(ctx context.Context, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 1024)
				buf = buf[:runtime.Stack(buf, false)] // 跟踪栈长度
				logger.Error(fmt.Sprintf("[PANIC]%v\n%s\n", r, buf))
				err = errors.New(fmt.Sprint(r))
			}
		}()

		return handler(ctx, req)
	}
}
