package environment

import "context"

// GetServerPort returns the port the service will run on
func GetGroupServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "GROUP_SERVER_PORT")
}

func GetGroupDbUser(ctx context.Context) string {
	return MustLookupString(ctx, "GROUP_DB_USER")
}

func GetGroupDbPassword(ctx context.Context) string {
	return MustLookupString(ctx, "GROUP_DB_PASSWORD")
}

func GetGroupDbHost(ctx context.Context) string {
	return MustLookupString(ctx, "GROUP_DB_HOST")
}

func GetGroupDbPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "GROUP_DB_PORT")
}

func GetGroupDbName(ctx context.Context) string {
	return MustLookupString(ctx, "GROUP_DB_NAME")
}

// GetServerPort returns the port the service will run on
func GetReflectionServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "REFLECTION_SERVER_PORT")
}

// GetGlobalDomain returns the infrastructure's global domain which can be used for various purposes like error details
func GetGlobalDomain(ctx context.Context) string {
	return MustLookupString(ctx, "GLOBAL_DOMAIN")
}

// GetDBSelectErrorReason returns the error reason that a GET request to the K8s API failed in UPPER_SNAKE_CASE
func GetDBSelectErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_SELECT_ERROR_REASON")
}

// GetNatsServerHost returns the host address of the NATS server
func GetNatsServerHost(ctx context.Context) string {
	return MustLookupString(ctx, "NATS_SERVER_HOST")
}

// GetNatsServerPort returns the host port of the NATS server
func GetNatsServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "NATS_SERVER_PORT")
}

func GetTraceCollectorHost(ctx context.Context) string {
	return MustLookupString(ctx, "TRACE_COLLECTOR_HOST")
}

func GetTraceCollectorPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "TRACE_COLLETOR_PORT")
}

// TODO: as env variable
// GroupCreationRequested is the name of the subject group-creation-requested events are published on
var GroupCreationRequested string = "group.GroupCreationRequested"

// TODO: as env variable
// HttpStatusCodeKey is the header key used internally to modify the http status code as suggested here: https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/
var HttpStatusCodeKey string = "x-http-code"
