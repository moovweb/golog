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

func NewConsoleDispatcher() *LogDispatcher {
	return NewLogDispatcher(os.Stdout)
}

func NewNullDispatcher() *LogDispatcher {
	return NewLogDispatcher(ioutil.Discard)
}

// Currently we do not support rolling logs, but this could be fixed by simply
// implementing a new io.Writer object for files which will perform the 
// rolling and use that writer in here instead of os.OpenFile(...)
const filePerms = 0644 // rw user, r everyone else
func NewFileDispatcher(filename string) (*LogDispatcher, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerms)
	if err != nil {
		return nil, err
	}
	return NewLogDispatcher(f), nil
}

