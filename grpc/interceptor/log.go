package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nioliu/commons/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"time"
)

func GetCallLogFunc() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		incomingContext, _ := metadata.FromIncomingContext(ctx)

		p, b := peer.FromContext(ctx)
		var remoteIp string
		var remoteProtocol string
		if b {
			remoteIp = p.Addr.String()
			remoteProtocol = p.Addr.Network()
		}

		before := time.Now()
		resp, err = handler(ctx, req)
		after := time.Now()
		duration := after.Sub(before).Microseconds()

		reqBytes, _ := json.Marshal(req)
		rspBytes, _ := json.Marshal(resp)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		infos := map[string]interface{}{
			"RemoteIp":       remoteIp,
			"RemoteProtocol": remoteProtocol,
			"Req":            string(reqBytes),
			"FullMethod":     info.FullMethod,
			"Server":         info.Server,
			"Resp":           string(rspBytes),
			"Error":          errStr,
			"Duration":       fmt.Sprintf("%dms", duration),
		}

		for k, v := range incomingContext {
			infos[k] = v
		}

		infoJson, _ := json.Marshal(infos)

		log.InfoWithCtxFields(ctx, "CallLog", zap.String("info", string(infoJson)))
		return resp, err
	}
}

// GetBackCallLogFunc Stdout log and customer fields, key is from ctx, will do ctx.Value(key...)
func GetBackCallLogFunc() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		before := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Now().Sub(before)

		reqBytes, _ := json.Marshal(req)
		rspBytes, _ := json.Marshal(reply)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		infos := map[string]interface{}{
			"Req":      string(reqBytes),
			"target":   cc.Target(),
			"Resp":     string(rspBytes),
			"Error":    errStr,
			"Duration": fmt.Sprintf("%dms", duration),
		}

		// metadata
		md, b := metadata.FromOutgoingContext(ctx)
		if b {
			for k, v := range md {
				infos[k] = v
			}
		}

		jsonInfo, _ := json.Marshal(infos)

		log.InfoWithCtxFields(ctx, "BackCallLog", zap.String("info", string(jsonInfo)))
		return err
	}
}
