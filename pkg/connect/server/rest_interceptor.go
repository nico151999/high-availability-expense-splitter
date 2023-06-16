package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"google.golang.org/protobuf/proto"
)

// httpResponseCodeModifier checks if the gRPC response contains an http status code header and sets the http status code accordingly
func httpResponseCodeModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set http status code
	httpStatusCodeKey := environment.HttpStatusCodeKey
	if vals := md.HeaderMD.Get(httpStatusCodeKey); len(vals) == 1 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, httpStatusCodeKey)
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	}

	return nil
}
