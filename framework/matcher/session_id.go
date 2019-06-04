package matcher

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/retro"
)

// SessionID matcher is for matching against SessionIDs in
// checkpoints. It allows the look-up of entities touched
// in checkpoints on a given session ID.
type sessionIDMatcher retro.SessionID

// DoesMatch will return (false, nil) if the given entity is not
// castable to a Checpoint.
//
// If the given entity is a checkpoint the session header field will
// be checked against the name
func (sidm sessionIDMatcher) DoesMatch(i interface{}) (bool, error) {

	fmt.Println(reflect.TypeOf(i))

	return false, nil
}

func NewSessionID(sid string) retro.Matcher {
	return sessionIDMatcher(retro.SessionID(sid))
}
