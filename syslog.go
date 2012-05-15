package golog

import (
	"net"
	"os"
	"fmt"
	"errors"
	"time"
)

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

type Facility int

const (
	LOCAL0 = iota + 16
	LOCAL1
	LOCAL2
	LOCAL3
	LOCAL4
	LOCAL5
	LOCAL6
	LOCAL7
)

func connectToSyslog() (sock net.Conn, err error) {
	logTypes := []string { "unixgram", "unix" }
	logPaths := []string { "/dev/log", "/var/run/syslog" }

	for _, network := range logTypes {
		for _, path := range logPaths {
			sock, err = net.Dial(network, path)
			if err == nil {
				fmt.Fprintf(os.Stderr, "syslog uses %s:%s\n", network, path)
				return sock, nil
			}
		}
	}
	return nil, err
	
}

type SyslogProcessor struct {
	priority Priority
	facility Facility
	dispatcher *LogDispatcher
}

func NewSyslogProcessor(f Facility, p Priority) (LogProcessor, error) {
	sw, err := connectToSyslog()
	if err != nil {
		errMsg := fmt.Sprintf("Error in NewSyslogProcessor: %s", err.Error())
		return nil, errors.New(errMsg)
	}
	
	dsp := NewLogDispatcher(sw)
	return &SyslogProcessor { facility: f, priority: p, dispatcher: dsp }, nil
}

const syslogMsgFormat = "<%d>%s %s: %s\n"
func (su *SyslogProcessor) Process(entry *LogEntry) {
	if entry.priority <= su.priority {
		key := (int(su.facility) * 8) + int(entry.priority)
		timestr := time.Unix(entry.created.Unix(), 0).UTC().Format(time.RFC3339)
		prefix := entry.prefix
		msg := fmt.Sprintf(syslogMsgFormat, key, timestr, prefix, entry.msg)
		su.dispatcher.Send(msg)
	}
}
