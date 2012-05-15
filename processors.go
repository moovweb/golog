package golog

import "sync"

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
}

type DefaultProcessor struct {
	mu sync.RWMutex							// Read/Write Lock used to protect the priority.
	priority Priority						// Messages need to be at least this important to get through.
	dispatcher *LogDispatcher		// Dispatcher used to send messages to the channel
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
	if entry.priority <= df.GetPriority() {
		msg := entry.priority.String() + ": " + entry.prefix + entry.msg
		df.dispatcher.Send(msg)
	}
}

// Initializers for LogProcessor
//
func NewProcessor(priority Priority, dispatcher *LogDispatcher) LogProcessor {
	return &DefaultProcessor { priority: priority, dispatcher: dispatcher }
}

func NewConsoleProcessor(priority Priority) LogProcessor {
	console := NewConsoleDispatcher()
	return NewProcessor(priority, console)
}

func NewFileProcessor(priority Priority, filename string) (LogProcessor, error) {
	filer, err := NewFileDispatcher(filename)
	if err != nil {
		return nil, err
	}
	return NewProcessor(priority, filer), nil
}


