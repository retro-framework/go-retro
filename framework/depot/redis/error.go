package redis

import "fmt"

type Error struct {
	Op  string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("redisdepot: op: %q err: %q", e.Op, e.Err)
}
