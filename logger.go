package main

type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}

// logger should never "fatal", warnings are either
// info or error, warning is not actionable enough
type NoopLogger struct{}

func (nl NoopLogger) Debug(...interface{})          {}
func (nl NoopLogger) Debugf(string, ...interface{}) {}
func (nl NoopLogger) Info(...interface{})           {}
func (nl NoopLogger) Infof(string, ...interface{})  {}
func (nl NoopLogger) Error(...interface{})          {}
func (nl NoopLogger) Errorf(string, ...interface{}) {}
