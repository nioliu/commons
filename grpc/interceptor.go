package grpc

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var logger, _ = zap.NewDevelopment()

func CallLog(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		"ctx":            ctx,
		"Req":            req,
		"FullMethod":     info.FullMethod,
		"Server":         info.Server,
		"Resp":           resp,
		"Error":          err,
	}

	logger.Info("CallLog", zap.Any("info", infos))
	return resp, err
}

func BackCallLog(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)

	infos := map[string]interface{}{
		"Req":    req,
		"target": cc.Target(),
		"Resp":   reply,
		"Error":  err,
	}

	logger.Info("BackCallLog", zap.Any("info", infos))
	return err
}
