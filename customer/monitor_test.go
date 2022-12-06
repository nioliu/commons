package customer

import (
	"context"
	"encoding/json"
	"github.com/nioliu/commons/component"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
	"time"
)

type Log struct {
	Name    string    `json:"name"`
	Time    time.Time `json:"time"`
	Country string    `json:"country"`
	Random  string    `json:"random"`
}

func TestInitMonitorClient(t *testing.T) {
	client, err := InitMonitorClient(context.Background(),
		"localhost:9009", "monitor_test", nil,
		WithDiaOpts(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 400000; i++ {
		id := component.CreateSnowflakeId("0118")
		l := &Log{
			Name:    "nioliu",
			Time:    time.Now(),
			Country: "CN",
			Random:  id,
		}
		bytes, err := json.Marshal(l)
		if err != nil {
			log.Fatalln(err)
		}
		if err = client.Send(bytes); err != nil {
			log.Fatalln(err)
		}
	}

	time.Sleep(time.Second * 10)
}
