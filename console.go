// Log Processor for outputting into the console.
//
package golog

import "os"

func NewConsoleProcessor(priority Priority, verbose bool) LogProcessor {
	return NewProcessorFromWriter(priority, os.Stdout, verbose)
}
