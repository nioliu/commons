package interceptor

import (
	"context"
	"github.com/nioliu/commons/component"
	"github.com/nioliu/commons/errs"
	"github.com/nioliu/commons/grpc/object"
	"github.com/nioliu/commons/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"net/http"
	"net/http/httputil"
	"time"
)

func PreActionForGolang(ctx context.Context, req *http.Request, now time.Time) (context.Context, *errs.ErrRsp, int) {
	var errRsp *errs.ErrRsp
	dumpRequest, err := httputil.DumpRequest(req, true)
	if err == nil {
		log.InfoWithCtxFields(ctx, "receive request", zap.String("request content", string(dumpRequest)))
	}

	// fill metadata
	m := &metadata.MD{}
	if err = object.SetRecMsgSecondTimeToMd(m, now.Unix()); err != nil {
		log.ErrorWithCtxFields(ctx, "set time to md failed", zap.Error(err))

		errRsp = &errs.ErrRsp{Code: 1, Description: "internal service error"}
		return ctx, errRsp, http.StatusInternalServerError
	}

	if err = object.SetRecMsgMilliSecondTimeToMd(m, now.UnixMilli()); err != nil {
		log.ErrorWithCtxFields(ctx, "set milli sec time failed", zap.Error(err))
		errRsp = &errs.ErrRsp{Code: 1, Description: "internal service error"}
		return ctx, errRsp, http.StatusInternalServerError
	}

	ctx = metadata.NewOutgoingContext(ctx, *m) // out stac
	traceId := GetTraceId(req)
	ctx = context.WithValue(ctx, "trace_id", traceId)
	regionData := GetReginData(req)
	ctx = context.WithValue(ctx, "region_data", regionData)

	return ctx, nil, http.StatusOK
}

func GetTraceId(r *http.Request) string {
	traceId := r.Header.Get("X-Trace-Id")
	if traceId == "" {
		traceId = component.CreateSnowflakeId("0")
	}

	return traceId
}

func GetReginData(r *http.Request) string {
	regionData := r.Header.Get("X-Region-Data")

	return regionData
}
