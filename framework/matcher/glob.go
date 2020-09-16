package matcher

import (
	"fmt"
	"reflect"

	"github.com/zyedidia/glob"
	"golang.org/x/xerrors"
)

func NewGlobPattern(pattern string) globPatternMatcher {
	glob, _ := glob.Compile(pattern)
	return globPatternMatcher{glob}
}

// TODO: test me one way or another (can't rely on upstream lib globbing without local tests)
type globPatternMatcher struct {
	glob *glob.Glob
}

func (gpm globPatternMatcher) DoesMatch(i interface{}) (Result, error) {
	if gpm.glob == nil {
		return ResultNoMatch(), xerrors.New("has no glob, possibly the pattern would not compile")
	}

	fmt.Println(reflect.TypeOf(i))

	if s, ok := i.(string); ok {
		return ResultBoolean(gpm.glob.Match([]byte(s))), nil
	}

	// if b, ok := i.([]byte); ok {
	// 	return ResultBoolean(gpm.glob.Match(b)), nil
	// }
	return ResultNoMatch(), xerrors.Errorf("don't know how to handle type in glob matcher")

	// switch m := i.(type) {
	// case string:
	// // TODO: how to do this and avoid an import cycle...
	// // case retro.PartitionName:
	// // 	return matcher.ResultSuccess(gpm.glob.Match([]byte(m))), nil
	// default:
	// }
}
