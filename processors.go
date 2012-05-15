package golog

import "sync"

type LogProcessor interface {
	GetPriority() Priority
	SetPriority(Priority)
	Process(*LogEntry)
}

type DefaultProcessor struct {
	mu sync.RWMutex
	priority Priority
	dispatcher *LogDispatcher
}

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


