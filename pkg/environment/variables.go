package environment

import (
	"context"
	"fmt"
)

// GetServerPort returns the port the service will run on
func GetGroupServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "GROUP_SERVER_PORT")
}

func GetDbUser(ctx context.Context) string {
	return MustLookupString(ctx, "DB_USER")
}

func GetDbPassword(ctx context.Context) string {
	return MustLookupString(ctx, "DB_PASSWORD")
}

func GetDbHost(ctx context.Context) string {
	return MustLookupString(ctx, "DB_HOST")
}

func GetDbPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "DB_PORT")
}

func GetDbName(ctx context.Context) string {
	return MustLookupString(ctx, "DB_NAME")
}

// GetServerPort returns the port the service will run on
func GetReflectionServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "REFLECTION_SERVER_PORT")
}

// GetGlobalDomain returns the infrastructure's global domain which can be used for various purposes like error details
func GetGlobalDomain(ctx context.Context) string {
	return MustLookupString(ctx, "GLOBAL_DOMAIN")
}

// GetDBSelectErrorReason returns the error reason that a DB select to the database failed in UPPER_SNAKE_CASE
func GetDBSelectErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_SELECT_ERROR_REASON")
}

// GetDBInsertErrorReason returns the error reason that a DB insert to the database failed in UPPER_SNAKE_CASE
func GetDBInsertErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_INSERT_ERROR_REASON")
}

// GetDBDeleteErrorReason returns the error reason that a DB delete to the database failed in UPPER_SNAKE_CASE
func GetDBDeleteErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_DELETE_ERROR_REASON")
}

// GetDBUpdateErrorReason returns the error reason that a DB update to the database failed in UPPER_SNAKE_CASE
func GetDBUpdateErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_UPDATE_ERROR_REASON")
}

// GetTaskPublicationErrorReason returns the error reason that a task could not be published in UPPER_SNAKE_CASE
func GetTaskPublicationErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "TASK_PUBLICATION_ERROR_REASON")
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
	return MustLookupUint16(ctx, "TRACE_COLLECTOR_PORT")
}

// TODO: as env variable with %s parameter
// GetGroupCreatedSubject returns the name of the subject events are published on when a group was created
func GetGroupCreatedSubject(groupId string) string {
	return fmt.Sprintf("%s.%s.groupCreated", GetGroupSubject(), groupId)
}

// TODO: as env variable
// GetGroupCreatedSubject returns the name of the subject group events are published on
func GetGroupSubject() string {
	return "group"
}

// TODO: as env variable
// HttpStatusCodeKey is the header key used internally to modify the http status code as suggested here: https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/
var HttpStatusCodeKey string = "x-http-code"
