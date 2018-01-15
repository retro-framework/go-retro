package resolver

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/framework/in-memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type OneEvent struct{}
type OtherEvent struct{}
type ExtraEvent struct{}

type dummyAggregate struct {
	seenEvents []types.Event
}

func (da *dummyAggregate) ReactTo(ev types.Event) error {
	da.seenEvents = append(da.seenEvents, ev)
	return nil
}

type dummySession struct{}

func (_ *dummySession) ReactTo(types.Event) error {
	return nil
}

type dummyCmd struct {
	s          *dummyAggregate
	wasApplied bool
}

func (dc *dummyCmd) SetState(s types.Aggregate) error {
	if agg, ok := s.(*dummyAggregate); ok {
		dc.s = agg
		return nil
	} else {
		return errors.New("can't cast aggregate state")
	}
}

func (dc *dummyCmd) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	if len(dc.s.seenEvents) != 2 {
		return nil, errors.New(fmt.Sprintf("can't apply ExtraEvent to dummyAggregate unless it has seen precisely two events so far (has seen %d)", len(dc.s.seenEvents)))
	}
	dc.wasApplied = true
	return []types.Event{ExtraEvent{}}, nil
}

type otherDummyCmd struct {
	dummyCmd
	t     *testing.T
	state dummyAggregate
}

func (odc otherDummyCmd) SetState(agg types.Aggregate) error {
	if wa, ok := agg.(*dummyAggregate); ok {
		odc.state = *wa
		return nil
	} else {
		return errors.New("can't cast")
	}
}

func (odc *otherDummyCmd) Apply(_ context.Context, agg types.Aggregate, _ types.Depot) ([]types.Event, error) {
	if _, ok := agg.(*dummyAggregate); !ok {
		odc.t.Fatal("can't typecast aggregate (%q) to concrete dummyAggregate type", agg)
	}
	return nil, nil
}

func Test_Resolver_DoesNotResolveCmdToAggregateWithoutID(t *testing.T) {

	// Arrange
	var (
		emd = memory.NewEmptyDepot()

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()
		dCmd = &dummyCmd{}

		err error
	)

	aggm.Register("agg", &dummyAggregate{})
	cmdm.Register(&dummyAggregate{}, dCmd)

	var r = resolver{aggm: aggm, cmdm: cmdm}

	// Act
	_, err = r.Resolve(context.Background(), emd, []byte(`{"path":"agg", "name":"dummyCmd"}`))

	// Assert
	test.H(t).NotNil(err)
	if rErr, ok := err.(Error); !ok {
		t.Fatal("could not cast err to Error")
	} else {
		test.H(t).StringEql("parse-agg-path", rErr.Op)
		test.H(t).StringEql("agg path \"agg\" does not split into exactly two parts", rErr.Err.Error())
	}
}

func Test_Resolver_ResolveExistingCmdToExistingAggregateSuccessfully(t *testing.T) {

	// Arrange
	var (
		md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}

		err error
	)

	aggm.Register("agg", &dummyAggregate{})
	cmdm.Register(&dummyAggregate{}, dCmd)
	cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

	var (
		r   = resolver{aggm: aggm, cmdm: cmdm}
		res types.CommandFunc
	)

	// Act
	res, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

	// Assert
	test.H(t).IsNil(err)

	// Act
	newEvs, err := res(context.Background(), &dummySession{}, md)

	// Assert
	test.H(t).IsNil(err)
	test.H(t).IntEql(1, len(newEvs))
}

// Test_Resolver_ResolveExistingCmdToExistingAggregateSuccessfully is to test
// that when addressing a non-existant aggregate an empty instance is returned.
func Test_Resolver_ResolveExistingCmdToNonExistantAggregateSuccessfully(t *testing.T) {

	// Arrange
	var (
		md = memory.NewDepot(map[string][]types.Event{"agg/456": []types.Event{OneEvent{}, OtherEvent{}}})
		//                                                 ^^^

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}

		err error
	)

	aggm.Register("agg", &dummyAggregate{})
	cmdm.Register(&dummyAggregate{}, dCmd)
	cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

	var (
		r   = resolver{aggm: aggm, cmdm: cmdm}
		res types.CommandFunc
	)

	// Act
	res, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
	//                                                                  ^^^

	// Assert
	test.H(t).IsNil(err)

	// Act
	newEvs, err := res(context.Background(), &dummySession{}, md)

	// Assert
	test.H(t).NotNil(err) // dummyCmd throws error in case the aggregate has not !!= 2 events in the history
	test.H(t).IntEql(0, len(newEvs))

}

func Benchmark_Resolver_ResolveExistingCmdSuccessfully(b *testing.B) {

	b.ReportAllocs()

	// Arrange
	var (
		md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}
	)

	aggm.Register("agg", &dummyAggregate{})
	cmdm.Register(&dummyAggregate{}, dCmd)
	cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

	var (
		r = resolver{aggm: aggm, cmdm: cmdm}
	)

	for n := 0; n < b.N; n++ {
		// Act
		r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

	}
}
