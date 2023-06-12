package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"

	"github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	logginggrpc "github.com/nico151999/high-availability-expense-splitter/pkg/logging/grpc"
	"github.com/rotisserie/eris"
	"github.com/rs/cors"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Server struct {
	http.Server

	tracerProvider *sdktrace.TracerProvider
}

func (server *Server) Close(ctx context.Context) error {
	errGr, ctx := errgroup.WithContext(ctx)
	errGr.Go(server.Server.Close)
	errGr.Go(func() error {
		return server.tracerProvider.Shutdown(ctx)
	})
	return errGr.Wait()
}

// Serve serves the underlying server
func (server *Server) Serve(
	ctx context.Context,
	ln net.Listener,
) error {
	addr := ln.Addr().String()
	log := logging.FromContext(ctx)

	log.Info("serving Connect and gRPC-Gateway",
		logging.String("address", addr))
	if err := server.Server.Serve(ln); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("closed server")
		} else {
			return eris.Wrap(err, "failed to serve http")
		}
	}
	return nil
}

func Listen(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open a tcp socket on %s", addr)
	}
	return ln, nil
}

// ListenAndServe combines the actions of listening, creating a new server and serving
func ListenAndServe[CONNECT_HANDLER any](
	ctx context.Context,
	addr string,
	svc CONNECT_HANDLER,
	registerServiceHandler ServiceHandlerRegistrarFunc,
	createServiceHandler ServiceHandlerCreatorFunc[CONNECT_HANDLER],
	serviceName string,
	traceCollectorUrl string,
	corsPatterns []string,
	allowedCorsHeaders []string,
	allowedCorsMethods []string,
) (*Server, error) {
	ln, err := Listen(addr)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to listen on %s", addr)
	}
	spanExporter, err := createOtlpExporter(ctx, traceCollectorUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed creating OTLP span exporter")
	}
	server, err := NewServer(
		ctx,
		ln,
		svc,
		registerServiceHandler,
		createServiceHandler,
		serviceName,
		spanExporter,
		corsPatterns,
		allowedCorsHeaders,
		allowedCorsMethods,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create server")
	}
	if err := server.Serve(
		ctx,
		ln,
	); err != nil {
		return nil, eris.Wrap(err, "failed serving services")
	}
	return server, nil
}

func NewServer[CONNECT_HANDLER any](
	ctx context.Context,
	ln net.Listener,
	svc CONNECT_HANDLER,
	registerServiceHandler ServiceHandlerRegistrarFunc,
	createServiceHandler ServiceHandlerCreatorFunc[CONNECT_HANDLER],
	serviceName string,
	spanExporter sdktrace.SpanExporter,
	corsPatterns []string,
	allowedCorsHeaders []string,
	allowedCorsMethods []string,
) (*Server, error) {
	addr := ln.Addr().String()
	log := logging.FromContext(ctx).Named("Server")
	ctx = logging.IntoContext(ctx, log)

	tp, err := initTracer(ctx, serviceName, spanExporter)
	if err != nil {
		return nil, eris.Wrap(err, "tracer could not be initialised")
	}

	grpclog.SetLoggerV2(logginggrpc.NewGrpcLoggerV2(log))

	// Note: this will succeed asynchronously, once we've started the server below.
	var conn *grpc.ClientConn
	{
		var err error
		conn, err = grpc.DialContext(
			context.Background(),
			"dns:///"+addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, eris.Wrap(err, "failed to dial server")
		}
	}

	restMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(httpResponseCodeModifier),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true},
			},
		}),
	)
	if err := registerServiceHandler(context.Background(), restMux, conn); err != nil {
		return nil, eris.Wrap(err, "failed to register gateway")
	}

	grpcMux := http.NewServeMux()
	grpcMux.Handle(createServiceHandler(
		svc,
		connect.WithInterceptors(
			otelconnect.NewInterceptor(otelconnect.WithTracerProvider(tp)),
			connect.UnaryInterceptorFunc(unaryLogInterceptorFunc(ctx)),
			connect.UnaryInterceptorFunc(unaryValidateInterceptorFunc()),
		),
	))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if matched, _ := regexp.MatchString("^application/(?:grpc|connect).*$", r.Header.Get("content-type")); matched {
			log.Info("received grpc/connect request",
				logging.String("path", r.URL.Path),
				logging.String("method", r.Method))
			grpcMux.ServeHTTP(w, r)
		} else {
			log.Info("received REST request",
				logging.String("path", r.URL.Path),
				logging.String("method", r.Method))
			restMux.ServeHTTP(w, r)
		}
	})

	c := cors.New(cors.Options{
		AllowedOrigins:     corsPatterns,
		AllowedHeaders:     allowedCorsHeaders,
		AllowedMethods:     allowedCorsMethods,
		OptionsPassthrough: false,
	})

	return &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(c.Handler(mux), &http2.Server{}),
		},
		tracerProvider: tp,
	}, nil
}
