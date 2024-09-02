package object

import (
	"context"
	"google.golang.org/grpc/metadata"
	"log"
	"testing"
	"time"
)

func TestGetRecMsgSecondTimeFromCtx(t *testing.T) {
	m := &metadata.MD{}
	err := SetRecMsgSecondTimeToMd(m, time.Now().Unix())
	if err != nil {
		log.Fatalln(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), *m)
	fromCtx, err := GetRecMsgSecondTimeFromCtx(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	t.Log(fromCtx)
}

func TestApiKey(t *testing.T) {
	m := &metadata.MD{}
	err := SetInnerApiKeyToMd(m)
	if err != nil {
		log.Fatalln(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), *m)
	fromCtx, err := CheckInnerApiKey(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	t.Log(fromCtx)
}
