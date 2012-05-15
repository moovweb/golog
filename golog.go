package golog

import (
	"fmt"
	"time"
)

type LogEntry struct {
	prefix string
	priority Priority
	msg string
	created time.Time //time.Now()
}

type Logger struct {
	prefix string
	processors map[string]LogProcessor
}

func NewLogger() *Logger {
	return &Logger { prefix: "", processors: map[string]LogProcessor{} }
}


func (dl *Logger) AddProcessor(name string, processor LogProcessor) {
	dl.processors[name] = processor
}

func (dl *Logger) LogP(p Priority, prefix string, format string, args ... interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := &LogEntry {
		prefix: prefix,
		priority: p,
		msg: message,
		created: time.Now(),
	}

	for _, p := range dl.processors {
		p.Process(entry)
	}
}

func (dl *Logger) Log(p Priority, format string, args ... interface{}) {
	dl.LogP(p, dl.prefix, format, args...)
}

func (dl *Logger) Debug(format string, args ... interface{}) {
	dl.Log(LOG_DEBUG, format, args)
}

func (dl *Logger) Info(format string, args ... interface{}) {
	dl.Log(LOG_INFO, format, args)
}

func (dl *Logger) Notice(format string, args ... interface{}) {
	dl.Log(LOG_NOTICE, format, args)
}

func (dl *Logger) Warning(format string, args ... interface{}) {
	dl.Log(LOG_WARNING, format, args)
}

func (dl *Logger) Error(format string, args ... interface{}) {
	dl.Log(LOG_ERR, format, args)
}

func (dl *Logger) Critical(format string, args ... interface{}) {
	dl.Log(LOG_CRIT, format, args)
}

func (dl *Logger) Alert(format string, args ... interface{}) {
	dl.Log(LOG_ALERT, format, args)
}

func (dl *Logger) Emergency(format string, args ... interface{}) {
	dl.Log(LOG_EMERG, format, args)
}

