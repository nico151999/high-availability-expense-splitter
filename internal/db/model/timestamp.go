package model

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timestamp struct {
	timestamppb.Timestamp
}

var _ sql.Scanner = (*Timestamp)(nil)

func (ts *Timestamp) Scan(src interface{}) (err error) {
	var t time.Time
	switch src := src.(type) {
	case time.Time:
		t = src
	case nil:
		t = time.Time{}
	default:
		return UnsupportedDataTypeError{
			DataType: fmt.Sprintf("%T", src),
		}
	}
	ts.Timestamp = timestamppb.Timestamp{
		Seconds: int64(t.Unix()),
		Nanos:   int32(t.Nanosecond()),
	}
	return nil
}

var _ driver.Valuer = (*Timestamp)(nil)

func (ts *Timestamp) Value() (driver.Value, error) {
	return ts.AsTime(), nil
}

func NewTimestamp(timestamp *timestamppb.Timestamp) *Timestamp {
	return &Timestamp{
		Timestamp: timestamppb.Timestamp{
			Seconds: timestamp.GetSeconds(),
			Nanos:   timestamp.GetNanos(),
		},
	}
}

func (ts *Timestamp) IntoProtoTimestamp() *timestamppb.Timestamp {
	return &ts.Timestamp
}
