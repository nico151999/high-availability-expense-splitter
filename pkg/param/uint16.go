package param

import (
	"context"
	"flag"
	"strconv"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
)

var _ flag.Value = (*Uint16Param)(nil)

// Uint16Param is a uint parameter that allows to differentiate between being set and not
type Uint16Param struct {
	set   bool
	value uint16
}

func (sf *Uint16Param) Set(x string) error {
	val, err := strconv.ParseUint(x, 10, 16)
	if err != nil {
		return eris.Wrapf(err, "the string %s cannot be parsed into a uint16", x)
	}
	sf.value = uint16(val)
	sf.set = true
	return nil
}

func (sf *Uint16Param) String() string {
	return strconv.FormatUint(uint64(sf.value), 10)
}

func (sf *Uint16Param) IsSet() bool {
	return sf.set
}

func (sf *Uint16Param) Must(ctx context.Context) uint16 {
	if !sf.IsSet() {
		logging.FromContext(ctx).Panic("a required parameter was not passed")
	}
	return sf.value
}
