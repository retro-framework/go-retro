package storage

import (
	"golang.org/x/xerrors"
)

var (
	ErrUnknownRef         = xerrors.New("storage: ref unknown")
	ErrUnknownSymbolicRef = xerrors.New("storage: symbolic ref unknown")
)
