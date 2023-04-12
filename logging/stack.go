package logging

import (
	"runtime/debug"
	"strings"
)

type StackInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

func withStack(entry *Entry) *Entry {
	stackStr := string(debug.Stack())
	stacks := strings.Split(stackStr, "\n")
	stacks = append(stacks[:1], stacks[3:]...) //去掉debug.Stack
	newStacks := []string{}
	skip := false
	for _, str := range stacks {
		if skip && strings.Contains(str, ".go:") {
			skip = false
			continue
		}
		skip = false
		if strings.Contains(str, "logging.") { //去掉日志包
			skip = true
			continue
		}
		newStacks = append(newStacks, str)
	}
	s := strings.Join(newStacks, "\n")
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
