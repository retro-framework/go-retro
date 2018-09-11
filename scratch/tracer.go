// +build ignore

package main

type Span interface{}

type StartSpanOptions interface{}

type StartSpanOption interface {
	Apply(*StartSpanOptions)
}

type Tracer interface {
	StartSpan(operationName string, opts ...StartSpanOption) Span
}
