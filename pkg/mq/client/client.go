package client

import (
	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/serialization"
)

const PROTOBUF_ENCODER = "protobuf"

func init() {
	nats.RegisterEncoder(PROTOBUF_ENCODER, &serialization.ProtobufSerializer{})
}

func NewProtoMQClient(natsServer string) (*nats.EncodedConn, error) {
	nc, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	encodedClient, err := nats.NewEncodedConn(nc, PROTOBUF_ENCODER)
	if err != nil {
		return nil, err
	}
	return encodedClient, nil
}
