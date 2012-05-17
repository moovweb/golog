// Log Processor for outputting into the console.
//
package golog

import "os"

func NewConsoleProcessor(priority Priority) LogProcessor {
	return NewProcessorFromWriter(priority, os.Stdout)
}
