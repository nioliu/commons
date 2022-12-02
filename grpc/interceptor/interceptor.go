package interceptor

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func GetCallLogFunc(ctx context.Context, logger *zap.Logger, fields ...zap.Field) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		incomingContext, _ := metadata.FromIncomingContext(ctx)

		p, b := peer.FromContext(ctx)
		var remoteIp string
		var remoteProtocol string
		if b {
			remoteIp = p.Addr.String()
			remoteProtocol = p.Addr.Network()
		}

		resp, err = handler(ctx, req)

		infos := map[string]interface{}{
			"RemoteIp":       remoteIp,
			"RemoteProtocol": remoteProtocol,
			"ctx":            incomingContext,
			"Req":            req,
			"FullMethod":     info.FullMethod,
			"Server":         info.Server,
			"Resp":           resp,
			"Error":          err,
			"Fields":         fields,
		}

		infoJson, _ := json.Marshal(infos)

		logger.Info("CallLog", zap.String("info", string(infoJson)))
		return resp, err
	}
}

// GetBackCallLogFunc Stdout log and customer fields
func GetBackCallLogFunc(ctx context.Context, logger *zap.Logger, fields ...zap.Field) grpc.UnaryClientInterceptor {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		infos := map[string]interface{}{
			"Req":    req,
			"target": cc.Target(),
			"Resp":   reply,
			"Error":  err,
			"Fields": fields,
		}

		jsonInfo, _ := json.Marshal(infos)

		logger.Info("BackCallLog", zap.String("info", string(jsonInfo)))
		return err
	}
}
