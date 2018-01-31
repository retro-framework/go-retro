package framework

import (
	"fmt"
	"os"
)

type Noop struct{}

func (nl Noop) Debug(...interface{})          {}
func (nl Noop) Debugf(string, ...interface{}) {}
func (nl Noop) Info(...interface{})           {}
func (nl Noop) Infof(string, ...interface{})  {}
func (nl Noop) Warn(...interface{})           {}
func (nl Noop) Warnf(string, ...interface{})  {}
func (nl Noop) Error(...interface{})          {}
func (nl Noop) Errorf(string, ...interface{}) {}

type Stdout struct{}

func (s Stdout) Debug(args ...interface{})                  { fmt.Fprint(os.Stdout, args) }
func (s Stdout) Debugf(pattern string, args ...interface{}) { fmt.Fprintf(os.Stdout, pattern, args) }
func (s Stdout) Info(args ...interface{})                   { fmt.Fprint(os.Stdout, args) }
func (s Stdout) Infof(pattern string, args ...interface{})  { fmt.Fprintf(os.Stdout, pattern, args) }
func (s Stdout) Warn(args ...interface{})                   { fmt.Fprint(os.Stdout, args) }
func (s Stdout) Warnf(pattern string, args ...interface{})  { fmt.Fprintf(os.Stdout, pattern, args) }
func (s Stdout) Error(args ...interface{})                  { fmt.Fprint(os.Stdout, args) }
func (s Stdout) Errorf(pattern string, args ...interface{}) { fmt.Fprintf(os.Stdout, pattern, args) }
