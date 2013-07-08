package golog

import (
	"fmt"
	"io"
	"sync"
)

// ***************************************************************************
// LogProcessor interface defines the method that we expect all LogProcessors
// to have.  For most intents and purposes, a DefaultProcessor should suffice,
// however if special formatting is required, a new Processor could be made
// (see syslog.go).
//
// The LogProcessor also offers the ability to change its default Priority
// level at runtime using the SetPriority(...) method.  Implementing
// Processors need to make sure that SetPriority and GetPriority are
// thread safe.  Use the DefaultProcessor as an example.
//
type LogProcessor interface {
	GetPriority() Priority
	SetPriority(Priority)
	Process(*LogEntry)
	Close() error
}

type DefaultProcessor struct {
	mu         sync.RWMutex   // Read/Write Lock used to protect the priority.
	priority   Priority       // Messages need to be at least this important to get through.
	Dispatcher *LogDispatcher // Dispatcher used to send messages to the channel
	TimeFormat string         // Format string for time, if blank, we use a default.
	Verbose    bool
}

// Atomically set the new priority.  All accesses to priority need to be
// through GetPriority in order to maintain thread safety.
func (df *DefaultProcessor) SetPriority(p Priority) {
	p = BoundPriority(p)
	df.mu.Lock()
	df.priority = p
	df.mu.Unlock()
}

func (df *DefaultProcessor) GetPriority() Priority {
	df.mu.RLock()
	defer df.mu.RUnlock()
	return df.priority
}

func (df *DefaultProcessor) Process(entry *LogEntry) {
	if entry.Priority <= df.GetPriority() {
		time := entry.Created
		var timeStamp string
		if len(df.TimeFormat) == 0 {
			timeStamp = fmt.Sprintf("%s %d %02d:%02d:%02d ", time.Month().String()[0:3], time.Day(), time.Hour(), time.Minute(), time.Second())
		} else {
			timeStamp = time.Format(df.TimeFormat)
		}

		msg := ""
		if df.Verbose {
			msg += timeStamp + entry.Priority.String() + ": "
		}

		msg += entry.Prefix + entry.Msg
		df.Dispatcher.Send(msg)
	}
}

func (df *DefaultProcessor) Close() error {
	return df.Dispatcher.Close()
}

// Initializers for LogProcessor
//
func NewProcessor(priority Priority, dispatcher *LogDispatcher, verbose bool) LogProcessor {
	return &DefaultProcessor{priority: priority, Dispatcher: dispatcher, Verbose: verbose}
}

func NewProcessorFromWriter(priority Priority, writer io.WriteCloser, verbose bool) LogProcessor {
	d := NewLogDispatcher(writer)
	return NewProcessor(priority, d, verbose)
}

func NewProcessorWithTimeFormat(priority Priority, dispatcher *LogDispatcher, format string) LogProcessor {
	return &DefaultProcessor{
		priority:   priority,
		Dispatcher: dispatcher,
		Verbose:    true,
		TimeFormat: format,
	}
}
