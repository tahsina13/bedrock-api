package xerrors

import "errors"

var (
	// StorageErrNotFound is returned when a requested key does not exist in the store
	StorageErrNotFound = errors.New("key not found")
)
