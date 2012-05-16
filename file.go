package golog

import "io"

// Currently we do not support rolling logs, but this could be fixed by simply
// implementing a new io.Writer object for files which will perform the 
// rolling and use that writer in here instead of os.OpenFile(...)
const filePerms = 0644 // rw user, r everyone else
func openFile(filename string) (io.Writer, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerms)
	if err != nil {
		return nil, err
	}
	return f, nil
}


func NewFileProcessor(priority Priority, filename string) (LogProcessor, error) {
	w, err := openFile(filename)	
	if err != nil {
		return nil, err
	}
	filer := NewLogDispatcher(w)
	return NewProcessor(priority, filer), nil
}
