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
		p, b := peer.FromContext(ctx)
		var remoteIp string
		var remoteProtocol string
		if b {
			remoteIp = p.Addr.String()
			remoteProtocol = p.Addr.Network()
		}

		// set trace id into ctx
		fromIncomingContext, b2 := metadata.FromIncomingContext(ctx)
		if b2 {
			traceId := fromIncomingContext.Get("trace_id")
			if traceId != nil {
				traceIdStr := traceId[0]
				ctx = context.WithValue(ctx, "trace_id", traceIdStr)
			}
		}

		before := time.Now()
		resp, err = handler(ctx, req)
		duration := time.Now().Sub(before).Microseconds()

		reqBytes, _ := json.Marshal(req)
		rspBytes, _ := json.Marshal(resp)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		log.InfoWithCtxFields(ctx, "call",
			zap.String("req", string(reqBytes)),
			zap.String("rsp", string(rspBytes)),
			zap.String("remote_ip", remoteIp),
			zap.String("remote_protocol", remoteProtocol),
			zap.String("full_method", info.FullMethod),
			zap.String("error", errStr),
			zap.String("duration", fmt.Sprintf("%dms", duration)))
		return resp, err
	}
}

// GetBackCallLogFunc Stdout log and customer fields, key is from ctx, will do ctx.Value(key...)
func GetBackCallLogFunc() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		before := time.Now()

		// extract trace id to log
		traceId := ctx.Value("trace_id")
		if traceId != nil {
			traceIdStr, ok := traceId.(string)
			if ok {
				// insert trace id into outgoing metadata
				outgoingContext, b := metadata.FromOutgoingContext(ctx)
				if b {
					outgoingContext.Set("trace_id", traceIdStr)
				}
				// overwrite outgoing ctx
				ctx = metadata.NewOutgoingContext(ctx, outgoingContext)
			}
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Now().Sub(before)
		reqBytes, _ := json.Marshal(req)
		rspBytes, _ := json.Marshal(reply)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		log.InfoWithCtxFields(ctx, "backcall",
			zap.String("req", string(reqBytes)),
			zap.String("rsp", string(rspBytes)),
			zap.String("target", cc.Target()),
			zap.String("error", errStr),
			zap.String("duration", fmt.Sprintf("%dms", duration)),
		)

		return err
	}
}
