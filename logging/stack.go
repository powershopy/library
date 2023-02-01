package logging

import "runtime"

type StackInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

func withStack(entry *Entry) *Entry {
	const depth = 20
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	s := []StackInfo{}
	for {
		frame, more := frames.Next()
		s = append(s, StackInfo{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})
		if !more {
			break
		}
	}
	if entry == nil {
		return WithFields(map[string]interface{}{
			"stacks": s,
		})
	} else {
		return entry.WithFields(map[string]interface{}{
			"stacks": s,
		})
	}
}

func WithStack() *Entry {
	return withStack(nil)
}
