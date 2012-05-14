package golog

import (
	"io"
	"os"
)

type LogDispatcher struct {
	w io.Writer
	p Priority
	ch chan *LogEntry
}

func (lw *LogDispatcher) Send(priority Priority, p string) {
	entry := LogEntry { w: lw.w, msg: p }
	lw.ch <- &entry
}

func NewLogDispatcher(writer io.Writer) *LogDispatcher {
	lw := LogDispatcher{ w: writer, ch: logchan }
	return &lw
}

func NewConsoleDispatcher() *LogDispatcher {
	return NewLogDispatcher(os.Stdout)
}

func NewSyslogDispatcher(facility, priority int) *LogDispatcher {
	return nil
}
