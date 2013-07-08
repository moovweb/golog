// The golog package is a logging framework for the Go language based on the
// go routine and channel features of the language.  In essence, log
// messages sent to a logger are sent through a global channel where a
// go routine listens to and services each log write serially.
//
// As of right now, all messages are sent through a single logger channel,
// guranteeing serialization of log messages.  In the future, golog will
// switch to the model of having a channel and go routine per resource
// (such as file, network, console, etc...) so that writes to any single
// resource are serialized, but writes to different resources can be
// parallelized.
//
package golog

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ****************************************************************************
// Log Priority used to filter messages sent to the loggers.
//
type Priority int

const (
	log_DISABLE Priority = iota - 1

	// Using syslog standard priorities.
	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

func Priorities() []Priority {
	return []Priority{
		LOG_EMERG,
		LOG_ALERT,
		LOG_CRIT,
		LOG_ERR,
		LOG_WARNING,
		LOG_NOTICE,
		LOG_INFO,
		LOG_DEBUG}
}

// If a priority is out of bounds given any input, we'll simply
// truncate it to the closest valid priority.
func BoundPriority(priority Priority) Priority {
	if priority < log_DISABLE {
		priority = log_DISABLE
	} else if priority > LOG_DEBUG {
		priority = LOG_DEBUG
	}
	return priority
}

func (p Priority) String() string {
	switch p {
	case LOG_EMERG:
		return "EMERGENCY"
	case LOG_ALERT:
		return "ALERT"
	case LOG_CRIT:
		return "CRITICAL"
	case LOG_ERR:
		return "ERROR"
	case LOG_WARNING:
		return "WARNING"
	case LOG_NOTICE:
		return "NOTICE"
	case LOG_INFO:
		return "INFO"
	case LOG_DEBUG:
		return "DEBUG"
	case log_DISABLE:
		return "DISABLED"
	}
	return "UNKNOWN(" + strconv.Itoa(int(p)) + ")"
}

func ParsePriority(p string) Priority {
	p = strings.ToUpper(p)
	switch {
	case LOG_EMERG.String() == p:
		return LOG_EMERG
	case LOG_ALERT.String() == p:
		return LOG_ALERT
	case LOG_CRIT.String() == p:
		return LOG_CRIT
	case LOG_ERR.String() == p:
		return LOG_ERR
	case LOG_WARNING.String() == p:
		return LOG_WARNING
	case LOG_NOTICE.String() == p:
		return LOG_NOTICE
	case LOG_INFO.String() == p:
		return LOG_INFO
	case LOG_DEBUG.String() == p:
		return LOG_DEBUG
	}
	return log_DISABLE
}

// ****************************************************************************
// The actual Logger structure, methods, and components.
// The current Logger is implemented by holding a map of LogProcessors.
// A new Logger can be created by calling the NewLogger methods further below.
// A Logger will be initialized with zero processors to begin with, and thus,
// will need to have LogProcessors added to it with AddProcessor before it
// can begin logging.
//
type Logger struct {
	// prefix used to prepend to logs if no other prefix is supplied.
	prefix     string
	processors map[string]LogProcessor
	mu         sync.RWMutex // Read/Write Lock used to protect the prefix.
}

// Storage object used to pass the log data over to the Processor.
type LogEntry struct {
	Prefix   string    // Prefix to prepend to the log message.
	Priority Priority  // Priority of the log message.
	Msg      string    // The actual message payload
	Created  time.Time // Time this message was created.
}

func (dl *Logger) SetPrefix(newPrefix string) {
	dl.mu.Lock()
	dl.prefix = newPrefix
	dl.mu.Unlock()
}

// Set/Get the priority of the Processor with the given name.
// If no processor with the given name exists, we return an error.
func (dl *Logger) SetPriority(procName string, newPriority Priority) error {
	newPriority = BoundPriority(newPriority)
	proc := dl.processors[procName]
	if proc != nil {
		proc.SetPriority(newPriority)
		return nil
	}
	return errors.New("Couldn't find log processor with name '" + procName + "'")
}

func (dl *Logger) GetPriority(procName string) (Priority, error) {
	proc := dl.processors[procName]
	if proc != nil {
		return proc.GetPriority(), nil
	}
	return LOG_EMERG, errors.New("Coudln't find log processor with name '" + procName + "'")
}

func (dl *Logger) GetPriorities() map[string]Priority {
	pmap := map[string]Priority{}
	for name, proc := range dl.processors {
		pmap[name] = proc.GetPriority()
	}
	return pmap
}

func (dl *Logger) GetMaxPriority() Priority {
	max := log_DISABLE
	for _, proc := range dl.processors {
		if p := proc.GetPriority(); p > max {
			max = p
		}
	}
	return max
}

// Add processors to this logger with the given name.  Names need to be
// unique against all other processors.  If a name conflict arises, we
// simply override the old processor with the same name with the new one.
func (dl *Logger) AddProcessor(name string, processor LogProcessor) {
	if p := dl.processors[name]; p != nil {
		p.Close()
	}

	if processor == nil {
		// If we're setting it to nil, let's take that as deleting the key.
		delete(dl.processors, name)
	} else {
		dl.processors[name] = processor
	}
}

func (dl *Logger) DisableProcessor(name string) {
	dl.processors[name].SetPriority(log_DISABLE)
}

func (dl *Logger) Close() {
	for name, proc := range dl.processors {
		delete(dl.processors, name)
		if proc != nil {
			proc.Close()
		}
	}
}

// Begin Logging interface.  The following methods are used for logging
// messages to whatever processors this logger is associated with.
//
func (dl *Logger) Plogf(priority Priority, prefix string, format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	if len(message) == 0 || message[len(message)-1] != '\n' {
		message = message + "\n"
	}

	entry := &LogEntry{
		Prefix:   prefix,
		Priority: BoundPriority(priority),
		Msg:      message,
		Created:  time.Now(),
	}

	for _, p := range dl.processors {
		p.Process(entry)
	}
}

func (dl *Logger) Logf(p Priority, format string, args ...interface{}) {
	dl.mu.RLock()
	prefix := dl.prefix
	dl.mu.RUnlock()

	dl.Plogf(p, prefix, format, args...)
}

func (dl *Logger) Debugf(format string, args ...interface{}) {
	dl.Logf(LOG_DEBUG, format, args...)
}

func (dl *Logger) Infof(format string, args ...interface{}) {
	dl.Logf(LOG_INFO, format, args...)
}

func (dl *Logger) Noticef(format string, args ...interface{}) {
	dl.Logf(LOG_NOTICE, format, args...)
}

func (dl *Logger) Warningf(format string, args ...interface{}) {
	dl.Logf(LOG_WARNING, format, args...)
}

func (dl *Logger) Errorf(format string, args ...interface{}) {
	dl.Logf(LOG_ERR, format, args...)
}

func (dl *Logger) Criticalf(format string, args ...interface{}) {
	dl.Logf(LOG_CRIT, format, args...)
}

func (dl *Logger) Alertf(format string, args ...interface{}) {
	dl.Logf(LOG_ALERT, format, args...)
}

func (dl *Logger) Emergencyf(format string, args ...interface{}) {
	dl.Logf(LOG_EMERG, format, args...)
}

// Create a new empty Logger with the given prefix.
// The prefix will be prepended to every log message unless
// LogP(...) is used, in which case, the prefix supplied by the 'prefix'
// parameter will be used instead.
//
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix, processors: map[string]LogProcessor{}}
}

// ****************************************************************************
// The following section covers the part of the Logger which listens to the
// channel for new messages to write.  All messages that are logged are
// eventually sent to this channel, and the following go routine services them
// to their appropriate Writer objects.
//
var logchan chan *LogMsg

const logQueueSize = 512

var die *int32 = new(int32)

func init() {
	logchan = make(chan *LogMsg, logQueueSize)
	go func() {
		for entry := range logchan {
			io.WriteString(entry.w, entry.msg)
			shouldDie := atomic.LoadInt32(die)
			if shouldDie > 0 {
				break
			}
		}
	}()
}

func FlushLogsAndDie() {
	atomic.AddInt32(die, 1)
	for i := 0; i < logQueueSize; i++ {
		select {
		case entry := <-logchan:
			io.WriteString(entry.w, entry.msg)
		default:
			break
		}
	}
}
