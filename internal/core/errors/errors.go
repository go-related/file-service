package errors

import "errors"

var (
	InvalidPortsInputs       = errors.New("invalid input, no ports to insert/update")
	TransactionAlreadyExists = errors.New("transaction already exists")
	TransactionRequeired     = errors.New("transaction is required, to perform this operation")
	NotFound                 = errors.New("item not found")
)
