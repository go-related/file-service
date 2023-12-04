package errors

import "errors"

var (
	InvalidPortsInputs = errors.New("invalid input, no ports to insert/update")
)
