package golog

import "io"

var logchan = make(chan *LogEntry, 10)

type LogEntry struct {
	w io.Writer
	msg string
}


func Init() {
	go func() {
		for entry := range(logchan) {
			io.WriteString(entry.w, entry.msg)
		}
	}()
}
