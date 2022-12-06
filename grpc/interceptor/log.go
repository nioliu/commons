package interceptor

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func GetCallLogFunc(ctx context.Context, logger *zap.Logger) grpc.UnaryServerInterceptor {
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
			"Req":            req,
			"FullMethod":     info.FullMethod,
			"Server":         info.Server,
			"Resp":           resp,
			"Error":          err,
		}

		for k, v := range incomingContext {
			infos[k] = v
		}

		infoJson, _ := json.Marshal(infos)

		logger.Info("CallLog", zap.String("info", string(infoJson)))
		return resp, err
	}
}

// GetBackCallLogFunc Stdout log and customer fields, key is from ctx, will do ctx.Value(key...)
func GetBackCallLogFunc(ctx context.Context, logger *zap.Logger, key ...string) grpc.UnaryClientInterceptor {
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
		}

		// customer key
		if key != nil {
			for _, k := range key {
				infos[k] = ctx.Value(k)
			}
		}

		// metadata
		md, b := metadata.FromOutgoingContext(ctx)
		if b {
			for k, v := range md {
				infos[k] = v
			}
		}

		jsonInfo, _ := json.Marshal(infos)

		logger.Info("BackCallLog", zap.String("info", string(jsonInfo)))
		return err
	}
}
