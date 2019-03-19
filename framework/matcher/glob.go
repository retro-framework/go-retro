package matcher

import (
	"github.com/zyedidia/glob"
	"golang.org/x/xerrors"

	"github.com/retro-framework/go-retro/framework/retro"
)

func NewGlobPattern(pattern string) retro.Matcher {
	glob, _ := glob.Compile(pattern)
	return globPatternMatcher{glob}
}

// TODO: test me one way or another (can't rely on upstream lib globbing without local tests)
type globPatternMatcher struct {
	glob *glob.Glob
}

func (gpm globPatternMatcher) DoesMatch(i interface{}) (bool, error) {
	if gpm.glob == nil {
		return false, xerrors.New("has no glob, possibly the pattern would not compile")
	}
	switch m := i.(type) {
	case string:
		return gpm.glob.Match([]byte(m)), nil
	case retro.PartitionName:
		return gpm.glob.Match([]byte(m)), nil
	default:
		return false, xerrors.Errorf("don't know how to handle type in glob matcher")
	}
}
