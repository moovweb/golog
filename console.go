package golog

import "os"

func NewConsoleProcessor(priority Priority) LogProcessor {
	console := NewLogDispatcher(os.Stdout)
	return NewProcessor(priority, console)
}
