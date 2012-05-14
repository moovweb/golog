package golog

import (
	"fmt"
	"errors"
	"time"
	"os"
)

type LogFormat interface {
	Log(*LogEntry)
}

type DefaultFormat struct {
	priority Priority
	dispatcher *LogDispatcher
}

func NewDefaultFormat() LogFormat {
	console := NewConsoleDispatcher()
	return &DefaultFormat { priority: LOG_DEBUG, dispatcher: console }
}

func (df *DefaultFormat) Log(entry *LogEntry) {
	if entry.priority <= df.priority {
		df.dispatcher.Send(entry.prefix + entry.msg)
	}
}


type SyslogFormat struct {
	priority Priority
	facility Facility
	dispatcher *LogDispatcher
}

func NewSyslogFormat(f Facility, p Priority) (LogFormat, error) {
	sw, err := NewSyslogWriter()
	if err != nil {
		errMsg := fmt.Sprintf("Error in NewSyslogFormat: %s", err.Error())
		return nil, errors.New(errMsg)
	}
	
	dsp := NewLogDispatcher(sw)
	return &SyslogFormat { facility: f, priority: p, dispatcher: dsp }, nil
}

func (su *SyslogFormat) Log(entry *LogEntry) {
	const syslogMsgFormat = "<%d>%s %s %s: %s\n"

	if entry.priority <= su.priority {
		key := (int(su.facility) * 8) + int(entry.priority)
		timestr := time.Unix(entry.created.Unix(), 0).UTC().Format(time.RFC3339)
		host, err := os.Hostname()
		if err != nil { host = "unknown" }
		prefix := entry.prefix
		msg := fmt.Sprintf(syslogMsgFormat, key, timestr, host, prefix, entry.msg)
		su.dispatcher.Send(msg)
	}
}

