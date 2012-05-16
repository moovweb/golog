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
	"time"
)

// ****************************************************************************
// Log Priority used to filter messages sent to the loggers.
//
type Priority int

const (
	// Using syslog standard priorities.
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
const log_DISABLE Priority = -1

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
	}
	return "UNKNOWN(" + strconv.Itoa(int(p)) + ")"
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
}

// Storage object used to pass the log data over to the Processor.
type LogEntry struct {
	prefix   string    // Prefix to prepend to the log message.
	priority Priority  // Priority of the log message.
	msg      string    // The actual message payload
	created  time.Time // Time this message was created.
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

// Add processors to this logger with the given name.  Names need to be 
// unique against all other processors.  If a name conflict arises, we
// simply override the old processor with the same name with the new one.
func (dl *Logger) AddProcessor(name string, processor LogProcessor) {
	dl.processors[name] = processor
}

func (dl *Logger) DisableProcessor(name string) {
	dl.processors[name].SetPriority(log_DISABLE)
}

// Begin Logging interface.  The following methods are used for logging
// messages to whatever processors this logger is associated with.
//
func (dl *Logger) LogP(priority Priority, prefix string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if len(message) == 0 || message[len(message)-1] != '\n' {
		message = message + "\n"
	}

	entry := &LogEntry{
		prefix:   prefix,
		priority: BoundPriority(priority),
		msg:      message,
		created:  time.Now(),
	}

	for _, p := range dl.processors {
		p.Process(entry)
	}
}

func (dl *Logger) Log(p Priority, format string, args ...interface{}) {
	dl.LogP(p, dl.prefix, format, args...)
}

func (dl *Logger) Debug(format string, args ...interface{}) {
	dl.Log(LOG_DEBUG, format, args...)
}

func (dl *Logger) Info(format string, args ...interface{}) {
	dl.Log(LOG_INFO, format, args...)
}

func (dl *Logger) Notice(format string, args ...interface{}) {
	dl.Log(LOG_NOTICE, format, args...)
}

func (dl *Logger) Warning(format string, args ...interface{}) {
	dl.Log(LOG_WARNING, format, args...)
}

func (dl *Logger) Error(format string, args ...interface{}) {
	dl.Log(LOG_ERR, format, args...)
}

func (dl *Logger) Critical(format string, args ...interface{}) {
	dl.Log(LOG_CRIT, format, args...)
}

func (dl *Logger) Alert(format string, args ...interface{}) {
	dl.Log(LOG_ALERT, format, args...)
}

func (dl *Logger) Emergency(format string, args ...interface{}) {
	dl.Log(LOG_EMERG, format, args...)
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

func init() {
	logchan = make(chan *LogMsg, logQueueSize)
	go func() {
		for entry := range logchan {
			io.WriteString(entry.w, entry.msg)
		}
	}()
}
