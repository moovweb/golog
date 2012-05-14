package golog

import (
	"io"
)

type LogEntry struct {
	w io.Writer
	msg []byte
}

type LogChanWriter struct {
	w io.Writer
	ch chan *LogEntry
}

func NewLogChanWriter(writer io.Writer) *LogChanWriter {
	lw := LogChanWriter{ w: writer, ch: logchan }
	return &lw
}

func (lw *LogChanWriter) Write(p []byte) (n int, err error) {
	entry := LogEntry { w: lw.w, msg: p }
	lw.ch <- &entry
	return len(p), nil
}

func (lw *LogChanWriter) WriteString(s string) (n int, err error) {
	return lw.Write([]byte(s))
}
