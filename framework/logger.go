package main

type Noop struct{}

func (nl Noop) Debug(...interface{})          {}
func (nl Noop) Debugf(string, ...interface{}) {}
func (nl Noop) Info(...interface{})           {}
func (nl Noop) Infof(string, ...interface{})  {}
func (nl Noop) Warn(...interface{})           {}
func (nl Noop) Warnf(string, ...interface{})  {}
func (nl Noop) Error(...interface{})          {}
func (nl Noop) Errorf(string, ...interface{}) {}
