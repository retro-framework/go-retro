package storage

import (
	"errors"
)

var (
	ErrUnknownRef         = errors.New("storage: ref unknown")
	ErrUnknownSymbolicRef = errors.New("storage: symbolic ref unknown")
)
