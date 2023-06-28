package server

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server/interceptor"
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
	server         http.Server
	tracerProvider *sdktrace.TracerProvider
}

func (server *Server) Shutdown(ctx context.Context) error {
	log := logging.FromContext(ctx).NewNamed("Shutdown")
	errGr, ctx := errgroup.WithContext(ctx)
	errGr.Go(func() error {
		if err := server.server.Shutdown(ctx); err != nil {
			msg := "failed shutting down server"
			log.Error(msg, logging.Error(err))
			return eris.Wrap(err, msg)
		}
		return nil
	})
	errGr.Go(func() error {
		if err := server.tracerProvider.Shutdown(ctx); err != nil {
			msg := "failed shutting down tracer provider"
			log.Error(msg, logging.Error(err))
			return eris.Wrap(err, msg)
		}
		return nil
	})
	return errGr.Wait()
}

// Serve serves the underlying server
func (server *Server) Serve(
	ctx context.Context,
	ln net.Listener,
) error {
	log := logging.FromContext(ctx).NewNamed("Serve")
	addr := ln.Addr().String()

	log.Info("serving Connect and gRPC-Gateway",
		logging.String("address", addr))

	serverResult := make(chan error)
	go func() {
		if err := server.server.Serve(ln); err != nil {
			if eris.Is(err, http.ErrServerClosed) || eris.Is(err, net.ErrClosed) {
				log.Info("closed server")
				serverResult <- nil
			} else {
				log.Error("failed to serve http", logging.Error(err))
				serverResult <- eris.Wrap(err, "failed to serve http")
			}
		}
	}()
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			msg := "failed shutting down server"
			log.Error(msg, logging.Error(err))
			return eris.Wrap(err, msg)
		}
	case err := <-serverResult:
		// shutdown other components as well now as the http server is closed
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			msg := "failed shutting down remaining server components"
			log.Error(msg, logging.Error(err))
			return eris.Wrap(err, msg)
		}
		return err
	}
	return nil
}

// Listen listens on a new socket. Note that this is a non-blocking call and the passed context has no effect on the socket.
func Listen(ctx context.Context, addr string) (net.Listener, error) {
	log := logging.FromContext(ctx).NewNamed("Listen")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("failed to open a tcp socket", logging.String("address", addr))
		return nil, eris.Wrapf(err, "failed to open a tcp socket on %s", addr)
	}
	return ln, nil
}

// ListenAndServe combines the actions of listening, creating a new server and serving. It shuts down the server when the context is closed.
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
) error {
	log := logging.FromContext(ctx).NewNamed("ListenAndServe")
	ctx = logging.IntoContext(ctx, log)

	ln, err := Listen(ctx, addr)
	if err != nil {
		return eris.Wrapf(err, "failed to listen on %s", addr)
	}
	spanExporter, err := createOtlpExporter(ctx, traceCollectorUrl)
	if err != nil {
		return eris.Wrap(err, "failed creating OTLP span exporter")
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
		return eris.Wrap(err, "failed to create server")
	}
	if err := server.Serve(
		ctx,
		ln,
	); err != nil {
		return eris.Wrap(err, "failed serving services")
	}
	return nil
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
	log := logging.FromContext(ctx).NewNamed("NewServer")
	ctx = logging.IntoContext(ctx, log)

	addr := ln.Addr().String()

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
		runtime.WithForwardResponseOption(interceptor.HttpResponseCodeModifier),
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
			connect.UnaryInterceptorFunc(interceptor.UnaryLogInterceptorFunc(ctx)),
			connect.UnaryInterceptorFunc(interceptor.UnaryValidateInterceptorFunc()),
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
		server: http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(c.Handler(mux), &http2.Server{}),
		},
		tracerProvider: tp,
	}, nil
}
