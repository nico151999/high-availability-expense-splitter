package param

import (
	"context"
	"flag"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

var _ flag.Value = (*StringParam)(nil)

// StringParam is a string parameter that allows to differentiate between being set and not
type StringParam struct {
	set   bool
	value string
}

func (sf *StringParam) Set(x string) error {
	sf.value = x
	sf.set = true
	return nil
}

func (sf *StringParam) String() string {
	return sf.value
}

func (sf *StringParam) IsSet() bool {
	return sf.set
}

func (sf *StringParam) Must(ctx context.Context) string {
	if !sf.IsSet() {
		logging.FromContext(ctx).Panic("a required parameter was not passed")
	}
	return sf.value
}
