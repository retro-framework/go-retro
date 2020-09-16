package matcher

import (
	"fmt"
	"reflect"
)

// SessionID matcher is for matching against SessionIDs in
// checkpoints. It allows the look-up of entities touched
// in checkpoints on a given session ID.
type sessionIDMatcher struct {
	pattern string
}

// DoesMatch will return (false, nil) if the given entity is not
// castable to a Checpoint.
//
// If the given entity is a checkpoint the session header field will
// be checked against the name
func (sidm sessionIDMatcher) DoesMatch(i interface{}) (Result, error) {
	fmt.Println(sidm.pattern, reflect.TypeOf(i))
	return ResultNoMatch(), nil
}

func NewSessionID(s string) sessionIDMatcher {
	return sessionIDMatcher{s}
}
