package matcher

// Result holds a result for a matc
// presently this is quite a simple type
// but I would hope to adorn it with other
// helpers such as maybe syntax highlighting
// matched results, or returning metadata
// about the match. TODO.
type Result struct {
	success   bool
	matchType Reason
	err       error
}

func (mr Result) Reason() Reason { return mr.matchType }
func (mr Result) Success() bool  { return mr.success }
func (mr Result) Failure() bool  { return !mr.Success() }

func ResultNoMatch() Result {
	return Result{success: false, matchType: ReasonUnknown}
}

func ResultSuccess() Result {
	return Result{success: true, matchType: ReasonUnknown}
}

func ResultBoolean(b bool) Result {
	return Result{success: b, matchType: ReasonUnknown}
}

func ResultAffixMatch(b bool) Result {
	return Result{success: b, matchType: ReasonAffix}
}

func ResultCheckpointMatch(b bool) Result {
	return Result{success: b, matchType: ReasonCheckpoint}
}
