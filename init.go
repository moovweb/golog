package golog

import "io"

var logchan = make(chan *LogMsg, 10)

type LogMsg struct {
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
