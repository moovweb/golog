package golog

import "io"

var logchan chan *LogMsg


const logQueueSize = 512
func Init() {
	logchan = make(chan *LogMsg, logQueueSize)
	go func() {
		for entry := range(logchan) {
			io.WriteString(entry.w, entry.msg)
		}
	}()
}

