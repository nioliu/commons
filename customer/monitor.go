package customer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nioliu/protocols/monitor"
	"google.golang.org/grpc"
)

type Monitor interface {
	Send(msg []byte) error
}

type MonitorClient struct {
	index string

	sendCli    monitor.MonitorService_SendClient
	receiveCli monitor.MonitorService_ReceiveClient

	// grpc options
	dialOpts        []grpc.DialOption
	sendCallOpts    []grpc.CallOption
	receiveCallOpts []grpc.CallOption
}

func (m *MonitorClient) Send(msg []byte) error {
	if m.sendCli == nil {
		return errors.New("send client is nil")
	}
	// check
	if !json.Valid(msg) {
		return errors.New("msg is not json type")
	}
	return m.sendCli.Send(&monitor.SendRequest{
		Msg:   msg,
		Index: m.index,
	})
}

// InitMonitorClient add is monitor service address, index is for es
func InitMonitorClient(ctx context.Context, add string, index string,
	in *monitor.ReceiveRequest, opts ...Option) (*MonitorClient, error) {

	m := &MonitorClient{index: index}
	apply(m, opts...)

	conn, err := grpc.DialContext(ctx, add, m.dialOpts...)
	if err != nil {
		return nil, err
	}

	client := monitor.NewMonitorServiceClient(conn)
	m.sendCli, err = client.Send(ctx, m.sendCallOpts...)
	if err != nil {
		return nil, err
	}

	if in == nil {
		in = &monitor.ReceiveRequest{Index: "default"}
	}

	m.receiveCli, err = client.Receive(ctx, in, m.receiveCallOpts...)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type Option func(client *MonitorClient)

func apply(client *MonitorClient, os ...Option) {
	for _, o := range os {
		o(client)
	}
}

func WithSendCallOpts(opts ...grpc.CallOption) Option {
	return func(client *MonitorClient) {
		client.sendCallOpts = opts
	}
}

func WithReceiveCallOpts(opts ...grpc.CallOption) Option {
	return func(client *MonitorClient) {
		client.receiveCallOpts = opts
	}
}

func WithDiaOpts(opts ...grpc.DialOption) Option {
	return func(client *MonitorClient) {
		client.dialOpts = opts
	}
}
