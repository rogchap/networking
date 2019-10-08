package main

import (
	"io"
	"os"
)

type tracer struct {
}

func (t *tracer) Init() {}

func (t *tracer) Ftrace(w io.Writer, route string) error {
	return nil
}

func (t *tracer) Trace(route string) error {
	return t.Ftrace(os.Stdout, route)
}
