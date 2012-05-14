package golog

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

type Writer struct {
	priority Priority
	facility Facility
}


