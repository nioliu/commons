package object

import (
	"context"
	"github.com/nioliu/commons/errs"
	"google.golang.org/grpc/metadata"
	"os"
	"strconv"
	"time"
)

type ContextKey string

const RecMsgSecondTimeKey = ContextKey("event_time_sec")
const RecMsgMilliSecondTimeKey = ContextKey("event_time_milli")
const TraceId = ContextKey("trace_id")
const InnerApiKey = ContextKey("inner_api_key")
const UserIdKey = ContextKey("user_id")
const ApiKeyName = "INNER_API_KEY"

func GetRecMsgSecondTimeFromCtx(ctx context.Context) (int64, error) {
	// get receive msg timestamp
	var recMsgTime int64
	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		return 0, errs.NewError(0, "can't find expected metadata info")
	}
	times := md[string(RecMsgSecondTimeKey)]
	if len(times) != 1 {
		return 0, errs.NewError(0, "unexpected time value from metadata")
	}
	recMsgTimeInt, err := strconv.Atoi(times[0])
	if err != nil {
		return 0, errs.NewError(0, "unexpected timestamp")
	}
	recMsgTime = int64(recMsgTimeInt)

	return recMsgTime, nil
}

func GetRecMsgMilliSecondTimeFromCtx(ctx context.Context) (int64, error) {
	// get receive msg timestamp
	var recMsgTime int64
	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		return 0, errs.NewError(0, "can't find expected metadata info")
	}
	times := md[string(RecMsgMilliSecondTimeKey)]
	if len(times) != 1 {
		return 0, errs.NewError(0, "unexpected time value from metadata")
	}
	recMsgTimeInt, err := strconv.Atoi(times[0])
	if err != nil {
		return 0, errs.NewError(0, "unexpected timestamp")
	}
	recMsgTime = int64(recMsgTimeInt)

	return recMsgTime, nil
}

func SetRecMsgSecondTimeToMd(md *metadata.MD, t int64) error {
	if md == nil {
		return errs.NewError(0, "metadata is nil")
	}
	if t == 0 {
		t = time.Now().Unix()
	}
	md.Append(string(RecMsgSecondTimeKey), strconv.Itoa(int(t)))
	return nil
}

func SetRecMsgMilliSecondTimeToMd(md *metadata.MD, t int64) error {
	if md == nil {
		return errs.NewError(0, "metadata is nil")
	}
	if t == 0 {
		t = time.Now().UnixMilli()
	}
	md.Append(string(RecMsgMilliSecondTimeKey), strconv.Itoa(int(t)))
	return nil
}

func SetInnerApiKeyToMd(md *metadata.MD) error {
	apiKey := os.Getenv(ApiKeyName)
	if apiKey == "" {
		return errs.NewError(0, "apikey is empty in env")
	}
	if md == nil {
		return errs.NewError(0, "metadata is nil")
	}
	md.Append(string(InnerApiKey), apiKey)
	return nil
}

func CheckInnerApiKey(ctx context.Context) (bool, error) {
	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		return false, errs.NewError(0, "can't find expected metadata info")
	}
	apiKey := md[string(InnerApiKey)]
	if len(apiKey) != 1 {
		return false, errs.NewError(0, "unexpected time value from metadata")
	}

	return os.Getenv(ApiKeyName) == apiKey[0], nil
}

func GetUserIDFromCtx(ctx context.Context) (string, error) {
	md, existed := metadata.FromIncomingContext(ctx)
	if existed && len(md.Get(string(UserIdKey))) >= 1 {
		return md.Get(string(UserIdKey))[0], nil
	}
	return "", errs.NewError(0, "can't find user id in anywhere")
}
