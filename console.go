// Log Processor for outputting into the console.
//
package golog

import "os"

func NewConsoleProcessor(priority Priority) LogProcessor {
	console := NewLogDispatcher(os.Stdout)
	return NewProcessor(priority, console)
}
