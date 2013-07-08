package golog

import (
	"io"
)

// Object which is sent through the log channel
type LogMsg struct {
	w   io.Writer // The writer we'll write msg to on the "other side"
	msg string    // Log message.
}

// ****************************************************************************
// The LogDispatcher will take incoming log messages, create LogMsg
// objects, and send them through the channel that it is associated with.
//
type LogDispatcher struct {
	w  io.WriteCloser // The resource Writer this dispatcher is associated with.
	ch chan *LogMsg   // The channel to send LogMsg objects to.
}

func (lw *LogDispatcher) Send(message string) {
	entry := LogMsg{w: lw.w, msg: message}
	lw.ch <- &entry
}

func (lw *LogDispatcher) Close() error {
	return lw.w.Close()
}

// Initializers of LogDispatcher
//
func NewLogDispatcher(writer io.WriteCloser) *LogDispatcher {
	lw := LogDispatcher{w: writer, ch: logchan}
	return &lw
}
