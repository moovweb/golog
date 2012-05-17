package golog

import (
	"errors"
	"fmt"
	"net"
	"os"
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

func SyslogFacilities() []Facility {
	return []Facility {
		LOCAL0,
		LOCAL1,
		LOCAL2,
		LOCAL3,
		LOCAL4,
		LOCAL5,
		LOCAL6,
		LOCAL7 }
}

// Create a socket connection to the syslog
func unixSyslog() (sock net.Conn, err error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog"}

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

func DialSyslog(network, raddr string) (sock net.Conn, err error) {
	if network == "" {
		return unixSyslog()
	}
	return net.Dial(network, raddr)
}

// ****************************************************************************
// The SyslogProcessor implements the LogProcessor interface.  It requires
// special formatting, thus why the DefaultProcessor could not be used.  It
// also needs to keep track of which facility we're logging to as we can have
// multiple sysloggers logging to different facilities.
//
type SyslogProcessor struct {
	*DefaultProcessor
	facility Facility
}

// Not only do we filter out messages whose priority is not high enough
// (DefaultProcessor behavior), but we also format the log message in a 
// special way using the priority and facility in a way that syslog 
// understand.
const syslogMsgFormat = "<%d>%s %s: %s"

func (su *SyslogProcessor) Process(entry *LogEntry) {
	if entry.Priority <= su.GetPriority() {
		key := (int(su.facility) * 8) + int(entry.Priority)
		priorityStr := entry.Priority.String()
		msg := entry.Prefix + entry.Msg
		msg = fmt.Sprintf(syslogMsgFormat, key, os.Args[0], priorityStr, msg)
		su.Dispatcher.Send(msg)
	}
}

// Initializer for the SyslogProcessor
//
func NewSyslogProcessorAt(network, addy string, f Facility, p Priority) (LogProcessor, error) {
	sw, err := DialSyslog(network, addy)
	if err != nil {
		errMsg := fmt.Sprintf("Error in NewSyslogProcessor: %s", err.Error())
		return nil, errors.New(errMsg)
	}

	dsp := NewLogDispatcher(sw)
	defaultProcessor := NewProcessor(p, dsp).(*DefaultProcessor)
	return &SyslogProcessor{ DefaultProcessor: defaultProcessor, facility: f }, nil
}

func NewSyslogProcessor(f Facility, p Priority) (LogProcessor, error) {
	return NewSyslogProcessorAt("", "", f, p)
}
