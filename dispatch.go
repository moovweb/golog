package golog

import (
	"io"
	"os"
)

type LogDispatcher struct {
	w io.Writer
	ch chan *LogEntry
}

func (lw *LogDispatcher) Send(message string) {
	entry := LogEntry { w: lw.w, msg: message }
	lw.ch <- &entry
}


func NewLogDispatcher(writer io.Writer) *LogDispatcher {
	lw := LogDispatcher{ w: writer, ch: logchan }
	return &lw
}

func NewConsoleDispatcher() *LogDispatcher {
	return NewLogDispatcher(os.Stdout)
}

