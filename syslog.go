package golog

import (
	"net"
	"os"
	"fmt"
	"errors"
	"sync"
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

func unixSyslog() (sock net.Conn, err error) {
	logTypes := []string { "unixgram", "unix" }
	logPaths := []string { "/dev/log", "/var/run/syslog" }
	
	for _, network := range logTypes {
		for _, path := range logPaths {
			sock, err = net.Dial(network, path)
			if err == nil {
				return sock, nil
			}
		}
	}
	return nil, err
	
}

type SyslogProcessor struct {
	mu sync.RWMutex
	priority Priority
	facility Facility
	dispatcher *LogDispatcher
}

func NewSyslogProcessor(f Facility, p Priority) (LogProcessor, error) {
	sw, err := unixSyslog()
	if err != nil {
		errMsg := fmt.Sprintf("Error in NewSyslogProcessor: %s", err.Error())
		return nil, errors.New(errMsg)
	}
	
	dsp := NewLogDispatcher(sw)
	return &SyslogProcessor { facility: f, priority: p, dispatcher: dsp }, nil
}

func (su *SyslogProcessor) SetPriority(priority Priority) {
	su.mu.Lock()
	su.priority = priority
	su.mu.Unlock()
}

func (su *SyslogProcessor) GetPriority() Priority {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.priority
}

const syslogMsgFormat = "<%d>%s: %s\n"
func (su *SyslogProcessor) Process(entry *LogEntry) {
	if entry.priority <= su.GetPriority() {
		key := (int(su.facility) * 8) + int(entry.priority)
		msg := entry.prefix + entry.msg
		msg = fmt.Sprintf(syslogMsgFormat, key, os.Args[0], msg)
		su.dispatcher.Send(msg)
	}
}

