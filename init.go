package golog

var logchan = make(chan *LogEntry, 10)

func Init() {
	go func() {
		for entry := range(logchan) {
			entry.w.Write(entry.msg)
		}
	}()
}
