package retro

// PatternMatcher defines a single function interface
// for matching patterns. It is used to compare the aggregate
// paths within an affix to the aggregate name being searched
// for. In a sane implementation it should support at least
// POSIX globbing and perhaps even Regular Expressions to
// allow for matching such as `users/*` or similar.
//
// In testing, this pattern matcher may be replaced with a
// no-op or static matcher.
type PatternMatcher interface {
	DoesMatch(pattern, partition string) (bool, error)
}
