package aggregates

import "github.com/leehambley/ls-cms/events"

type Aggregate interface {
	ReactTo(events.Event) error
}
