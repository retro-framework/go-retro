package depot

import (
	"github.com/pkg/errors"
	"github.com/zyedidia/glob"

	"github.com/retro-framework/go-retro/framework/matcher"
)

// TODO: test me one way or another (can't rely on upstream lib globbing without local tests)
type GlobPatternMatcher struct{}

func (_ GlobPatternMatcher) DoesMatch(pattern, partition string) (matcher.Result, error) {
	glob, err := glob.Compile(pattern)
	if err != nil {
		return matcher.ResultNoMatch(), errors.Wrap(err, "can't compile glob pattern")
	}
	return matcher.ResultBoolean(glob.MatchString(partition)), nil
}
