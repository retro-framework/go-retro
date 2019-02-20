package packing

import "golang.org/x/xerrors"

var (
	ErrAffixScan      = xerrors.New("packing: err scanning affix")
	ErrCheckpointScan = xerrors.New("packing: err scanning checkpoint")

	ErrInvalidPartitioName = xerrors.New("packing: invalid partition name")
)
