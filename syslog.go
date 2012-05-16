package golog

import (
	"net"
	"os"
	"fmt"
	"errors"
	"sync"
)

// Syslog facilities to log to.  We only list the LOCAL set as others
// are reserved for specific purposes.
type Facility int

const (
	LOCAL0 Facility = iota + 16
	LOCAL1
	LOCAL2
	LOCAL3
	LOCAL4
	LOCAL5
	LOCAL6
	LOCAL7
)

// Create a socket connection to the syslog
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

// ****************************************************************************
// The SyslogProcessor implements the LogProcessor interface.  It requires
// special formatting, thus why the DefaultProcessor could not be used.  It
// also needs to keep track of which facility we're logging to as we can have
// multiple sysloggers logging to different facilities.
//
type SyslogProcessor struct {
	mu sync.RWMutex
	priority Priority
	facility Facility
	dispatcher *LogDispatcher
}

// Atomically set the priority to the given value (adjusting for out of bounds)
func (su *SyslogProcessor) SetPriority(priority Priority) {
	priority = BoundPriority(priority)
	su.mu.Lock()
	su.priority = priority
	su.mu.Unlock()
}

// Get the priority via the protected read locks
func (su *SyslogProcessor) GetPriority() Priority {
	su.mu.RLock()
	defer su.mu.RUnlock()
	return su.priority
}

// Not only do we filter out messages whose priority is not high enough
// (DefaultProcessor behavior), but we also format the log message in a 
// special way using the priority and facility in a way that syslog 
// understand.
const syslogMsgFormat = "<%d>%s %s: %s"
func (su *SyslogProcessor) Process(entry *LogEntry) {
	if entry.priority <= su.GetPriority() {
		key := (int(su.facility) * 8) + int(entry.priority)
		priorityStr := entry.priority.String()
		msg := entry.prefix + entry.msg
		msg = fmt.Sprintf(syslogMsgFormat, key, os.Args[0], priorityStr, msg)
		su.dispatcher.Send(msg)
	}
}

// Initializer for the SyslogProcessor
//
func NewSyslogProcessor(f Facility, p Priority) (LogProcessor, error) {
	sw, err := unixSyslog()
	if err != nil {
		errMsg := fmt.Sprintf("Error in NewSyslogProcessor: %s", err.Error())
		return nil, errors.New(errMsg)
	}
	
	dsp := NewLogDispatcher(sw)
	return &SyslogProcessor { facility: f, priority: p, dispatcher: dsp }, nil
}

