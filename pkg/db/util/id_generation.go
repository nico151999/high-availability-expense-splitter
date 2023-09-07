package util

import (
	"fmt"
	"strconv"
	"time"
)

// GenerateIdWithPrefix generate an ID using the last 15 characters of the deciaml representation of the current unix timestamp prefixed with a keyword
func GenerateIdWithPrefix(prefix string) string {
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	return fmt.Sprintf("%s-%s", prefix, timestamp[len(timestamp)-15:])
}
