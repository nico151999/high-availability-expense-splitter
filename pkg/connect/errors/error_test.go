package errors_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestErrorDetail(t *testing.T) {
	tests := []struct {
		code                 connect.Code
		msg                  string
		details              []protoreflect.ProtoMessage
		expectDetailsToMatch bool
	}{
		{
			code: connect.CodeAborted,
			msg:  "the transaction was aborted",
			details: []protoreflect.ProtoMessage{
				&errdetails.ErrorInfo{
					Reason: "my-reason",
					Domain: "my-domain",
				},
			},
			expectDetailsToMatch: true,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		conErr := errors.NewErrorWithDetails(ctx, test.code, test.msg, test.details)
		if conErr.Code() != test.code {
			t.Errorf("expected code %d, got %d", test.code, conErr.Code())
		}
		if conErr.Message() != test.msg {
			t.Errorf("expected message %s, got %s", test.msg, conErr.Message())
		}
		t.Run("typed details", func(t *testing.T) {
			details, err := errors.TypedDetailsFromError[*errdetails.ErrorInfo](ctx, conErr, 0, -1)
			if err != nil {
				t.Error(err)
			}
			for i, detail := range details {
				if test.expectDetailsToMatch != cmp.Equal(detail, test.details[i], cmpopts.IgnoreUnexported(errdetails.ErrorInfo{})) {
					t.Errorf("the detail at index %d was unexpected", i)
				}
			}
		})
		t.Run("first detail", func(t *testing.T) {
			detail, err := errors.FirstTypedDetailFromError[*errdetails.ErrorInfo](ctx, conErr)
			if len(test.details) == 0 && err == nil {
				t.Error("expected to fail getting first detail since no detail was passed")
			} else if err != nil {
				t.Error(err)
			}
			if test.expectDetailsToMatch != cmp.Equal(detail, test.details[0], cmpopts.IgnoreUnexported(errdetails.ErrorInfo{})) {
				t.Errorf("the first detail was unexpected")
			}
		})
	}
}
