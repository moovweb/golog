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

type Logger interface {
	Logf(Priority, string, ... interface{})
	Debugf(string, ... interface{})
	Infof(string, ... interface{})
	Noticef(string, ... interface{})
	Warningf(string, ... interface{})
	Errorf(string, ... interface{})
	Criticalf(string, ... interface{})
	Alertf(string, ... interface{})
	Emergencyf(string, ... interface{})
}

type PrintLogger struct {}

func NewDefaultLogger() Logger {
	return &PrintLogger{}
}

func (pl *PrintLogger) Logf(p Priority, format string, args ... interface{}) {
	fmt.Printf(format, args)
}

func (pl *PrintLogger) Debugf(format string, args ... interface{}) {
	pl.Logf(LOG_DEBUG, format, args)
}

func (pl *PrintLogger) Infof(format string, args ... interface{}) {
	pl.Logf(LOG_INFO, format, args)
}

func (pl *PrintLogger) Noticef(format string, args ... interface{}) {
	pl.Logf(LOG_NOTICE, format, args)
}

func (pl *PrintLogger) Warningf(format string, args ... interface{}) {
	pl.Logf(LOG_WARNING, format, args)
}

func (pl *PrintLogger) Errorf(format string, args ... interface{}) {
	pl.Logf(LOG_ERR, format, args)
}

func (pl *PrintLogger) Criticalf(format string, args ... interface{}) {
	pl.Logf(LOG_CRIT, format, args)
}

func (pl *PrintLogger) Alertf(format string, args ... interface{}) {
	pl.Logf(LOG_ALERT, format, args)
}

func (pl *PrintLogger) Emergencyf(format string, args ... interface{}) {
	pl.Logf(LOG_EMERG, format, args)
}
