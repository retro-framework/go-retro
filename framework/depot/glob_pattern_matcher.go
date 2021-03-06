package depot

import (
	"github.com/pkg/errors"
	"github.com/zyedidia/glob"
)

// TODO: test me one way or another (can't rely on upstream lib globbing without local tests)
type GlobPatternMatcher struct{}

func (_ GlobPatternMatcher) DoesMatch(pattern, partition string) (bool, error) {
	glob, err := glob.Compile(pattern)
	if err != nil {
		return false, errors.Wrap(err, "can't compile glob pattern")
	}
	return glob.MatchString(partition), nil
}
