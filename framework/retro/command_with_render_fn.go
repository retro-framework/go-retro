package retro

import (
	"context"
	"io"
)

type CommandWithRenderFn interface {
	Render(context.Context, io.Writer, Session, CommandResult) error
}
