package environment

import (
	"context"
	"os"
	"strconv"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func MustLookupString(ctx context.Context, key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		logging.FromContext(ctx).Panic("failed looking up required environment variable", logging.String("envKey", key))
	}
	return val
}

func MustLookupUint16(ctx context.Context, key string) uint16 {
	val, exists := os.LookupEnv(key)
	if !exists {
		logging.FromContext(ctx).Panic("failed looking up required environment variable", logging.String("envKey", key))
	}
	parsedVal, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		logging.FromContext(ctx).Panic("failed parsing looked up environment variable as uint16", logging.String("envKey", key), logging.String("envValue", val))
	}
	return uint16(parsedVal)
}
