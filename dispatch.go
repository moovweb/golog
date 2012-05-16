package golog

import (
	"io"
	"io/ioutil"
	"os"
)

// Object which is sent through the log channel
type LogMsg struct {
	w io.Writer		// The writer we'll write msg to on the "other side"
	msg string		// Log message.
}

// ****************************************************************************
// The LogDispatcher will take incoming log messages, create LogMsg
// objects, and send them through the channel that it is associated with.
// 
type LogDispatcher struct {
	w io.Writer				// The resource Writer this dispatcher is associated with.
	ch chan *LogMsg		// The channel to send LogMsg objects to.
}

func (lw *LogDispatcher) Send(message string) {
	entry := LogMsg { w: lw.w, msg: message }
	lw.ch <- &entry
}

// Initializers of LogDispatcher
//
func NewLogDispatcher(writer io.Writer) *LogDispatcher {
	lw := LogDispatcher{ w: writer, ch: logchan }
	return &lw
}

func NewNullDispatcher() *LogDispatcher {
	return NewLogDispatcher(ioutil.Discard)
}

