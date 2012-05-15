package golog

import (
	"io"
	"io/ioutil"
	"os"
)

type LogMsg struct {
	w io.Writer
	msg string
}

type LogDispatcher struct {
	w io.Writer
	ch chan *LogMsg
}

func (lw *LogDispatcher) Send(message string) {
	entry := LogMsg { w: lw.w, msg: message }
	lw.ch <- &entry
}


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

const filePerms = 0644 // rw user, r everyone else
func NewFileDispatcher(filename string) (*LogDispatcher, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerms)
	if err != nil {
		return nil, err
	}
	return NewLogDispatcher(f), nil
}

