// create a channel writer and make sure that messages are
// being filtered at different levels.  Also use SetPriority and GetPriority

package golog

import (
	"strings"
	"testing"
	"time"
)

type ChanWriter struct {
	msg chan string
}

func (ch *ChanWriter) Write(b []byte) (n int, err error) {
	msg := string(b)
	ch.msg <- msg
	return len(b), nil
}

func (ch *ChanWriter) Close() error {
	close(ch.msg)
	return nil
}

func (ch *ChanWriter) CloseRenew() chan string {
	close(ch.msg)
	prev := ch.msg
	ch.msg = make(chan string, 31)
	return prev
}

func NewChanWriter() *ChanWriter {
	ch := make(chan string, 31)
	return &ChanWriter{msg: ch}
}

func checkFiltersForPriority(priority Priority, logger *Logger, chw *ChanWriter, t *testing.T) {
	logger.SetPriority("chan", priority)
	// Try to write a log of all priority levels, regardless of what our filter is
	for _, p := range Priorities() {
		logger.Logf(p, "Mmm, cherry blossom tea <3")
	}
	dur, _ := time.ParseDuration("100ms")
	time.Sleep(dur)
	receiver := chw.CloseRenew()

	for msg := range receiver {
		strPriority := strings.Split(msg, ":")[0]
		msgPriority := ParsePriority(strPriority)
		if msgPriority > priority {
			errmsg := "Received a message that should've been filtered!\n"
			errmsg += "Min Priority: %s\nMsg Priority: %s"
			t.Errorf(errmsg, priority.String(), strPriority)
		}
	}
}

func TestFilteredPriorities(t *testing.T) {
	priority := LOG_DEBUG
	chw := NewChanWriter()
	proc := NewProcessorFromWriter(priority, chw, true)
	logger := NewLogger("filter: ")
	logger.AddProcessor("chan", proc)

	for _, p := range Priorities() {
		checkFiltersForPriority(p, logger, chw, t)
	}
}
