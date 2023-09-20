package serialization

import (
	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ErrNoProtoMessage = eris.New("the passed message needs to be a protobuf message")

var _ nats.Encoder = (*ProtobufSerializer)(nil)

type ProtobufSerializer struct{}

func (ps *ProtobufSerializer) Encode(subject string, v any) ([]byte, error) {
	msg, err := checkProtoMessage(v)
	if err != nil {
		return nil, err
	}
	res, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ps *ProtobufSerializer) Decode(subject string, data []byte, vPtr any) error {
	target, err := checkProtoMessage(vPtr)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func checkProtoMessage(v any) (protoreflect.ProtoMessage, error) {
	msg, ok := v.(protoreflect.ProtoMessage)
	if !ok {
		return nil, ErrNoProtoMessage
	}
	return msg, nil
}
