package main

import (
	"context"
	"strings"

	"github.com/leehambley/ls-cms/aggregates"
	"github.com/leehambley/ls-cms/storage"
)

type Cmd struct {
	name string
	args ApplicationCmdArgs
}

func (c Cmd) Name() string             { return c.name }
func (c Cmd) Path() string             { return strings.Split(c.name, "/")[0] }
func (c Cmd) Args() ApplicationCmdArgs { return c.args }

type AggregateTransFunc func(context.Context, aggregates.Session, storage.Depot) ([]string, error)
type Command interface {
	Apply(context.Context, aggregates.Session, storage.Depot) ([]string, error)
}
