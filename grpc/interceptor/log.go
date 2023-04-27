package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
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
		duration := time.Now().Sub(before)

		reqBytes, _ := json.Marshal(req)
		rspBytes, _ := json.Marshal(resp)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		log.InfoWithCtxFields(ctx, "call",
			zap.ByteString("req", reqBytes),
			zap.ByteString("rsp", rspBytes),
			zap.String("remote_ip", remoteIp),
			zap.String("remote_protocol", remoteProtocol),
			zap.String("full_method", info.FullMethod),
			zap.String("error", errStr),
			zap.Duration("duration", duration))
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
		if reqByte, ok := req.([]byte); ok { // 转换byte
			reqByte = bytes.ReplaceAll(reqByte, []byte("\n"), []byte(""))
			reqBytes = reqByte
		}

		rspBytes, _ := json.Marshal(reply)
		if rspByte, ok := reply.([]byte); ok { // 转换byte
			rspByte = bytes.ReplaceAll(rspByte, []byte("\n"), []byte(""))
			rspBytes = rspByte
		}

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		log.InfoWithCtxFields(ctx, "backcall",
			zap.ByteString("req", reqBytes),
			zap.ByteString("rsp", rspBytes),
			zap.String("target", cc.Target()),
			zap.String("error", errStr),
			zap.Duration("duration", duration),
		)

		return err
	}
}
