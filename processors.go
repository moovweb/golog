package golog


type LogProcessor interface {
	Process(*LogEntry)
}

type DefaultProcessor struct {
	priority Priority
	dispatcher *LogDispatcher
}

func NewDefaultProcessor() LogProcessor {
	console := NewConsoleDispatcher()
	return &DefaultProcessor { priority: LOG_DEBUG, dispatcher: console }
}

func (df *DefaultProcessor) Process(entry *LogEntry) {
	if entry.priority <= df.priority {
		df.dispatcher.Send(entry.prefix + entry.msg)
	}
}

