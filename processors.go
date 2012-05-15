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

func NewDefaultProcessor() LogProcessor {
	console := NewConsoleDispatcher()
	return &DefaultProcessor { priority: LOG_DEBUG, dispatcher: console }
}

func (df *DefaultProcessor) SetPriority(p Priority) {
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
		df.dispatcher.Send(entry.prefix + entry.msg)
	}
}

