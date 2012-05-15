package golog

import (
	"fmt"
	"time"
	"strconv"
	"errors"
)

// some constants
type Priority int

const (
	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

func BoundPriority(priority Priority) Priority {
	if priority < LOG_EMERG {
		priority = LOG_EMERG
	} else if priority > LOG_DEBUG {
		priority = LOG_DEBUG
	}
	return priority
}

func (p Priority) String() string {
	switch p {
		case LOG_EMERG: return "EMERGENCY"
		case LOG_ALERT: return "ALERT"
		case LOG_CRIT: return "CRITICAL"
		case LOG_ERR: return "ERROR"
		case LOG_WARNING: return "WARNING"
		case LOG_NOTICE: return "NOTICE"
		case LOG_INFO: return "INFO"
		case LOG_DEBUG: return "DEBUG"
	}
	return "UNKNOWN(" + strconv.Itoa(int(p)) + ")"
}

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

func (dl *Logger) SetPriority(processorName string, newPriority Priority) error {
	newPriority = BoundPriority(newPriority)
	proc := dl.processors[processorName]
	if proc != nil {
		proc.SetPriority(newPriority)
		return nil
	}
	return errors.New("Couldn't find log processor with name '" + processorName + "'")
}

func (dl *Logger) AddProcessor(name string, processor LogProcessor) {
	dl.processors[name] = processor
}

func (dl *Logger) LogP(priority Priority, prefix string, format string, args ... interface{}) {
	message := fmt.Sprintf(format, args...)
	if len(message) == 0 ||  message[len(message)-1] != '\n' {
		message = message + "\n"
	}

	entry := &LogEntry {
		prefix: prefix,
		priority: BoundPriority(priority),
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
	dl.Log(LOG_DEBUG, format, args...)
}

func (dl *Logger) Info(format string, args ... interface{}) {
	dl.Log(LOG_INFO, format, args...)
}

func (dl *Logger) Notice(format string, args ... interface{}) {
	dl.Log(LOG_NOTICE, format, args...)
}

func (dl *Logger) Warning(format string, args ... interface{}) {
	dl.Log(LOG_WARNING, format, args...)
}

func (dl *Logger) Error(format string, args ... interface{}) {
	dl.Log(LOG_ERR, format, args...)
}

func (dl *Logger) Critical(format string, args ... interface{}) {
	dl.Log(LOG_CRIT, format, args...)
}

func (dl *Logger) Alert(format string, args ... interface{}) {
	dl.Log(LOG_ALERT, format, args...)
}

func (dl *Logger) Emergency(format string, args ... interface{}) {
	dl.Log(LOG_EMERG, format, args...)
}


func NewLogger(prefix string) *Logger {
	return &Logger { prefix: prefix, processors: map[string]LogProcessor{} }
}

