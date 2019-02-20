package depot

import "golang.org/x/xerrors"

var (
	ErrWritePacked = xerrors.New("depot: could not write packed object")
)
