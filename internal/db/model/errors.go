package model

import "fmt"

var _ error = (*UnsupportedDataTypeError)(nil)

type UnsupportedDataTypeError struct {
	DataType string
}

func (e UnsupportedDataTypeError) Error() string {
	return fmt.Sprintf("unsupported data type %s", e.DataType)
}
