// Log Processor for outputting into a file.
// Currently we do not support rolling logs, but this could be fixed by simply
// implementing a new io.Writer object for files which will perform the 
// rolling and use that writer in here instead of os.OpenFile(...)
//
package golog

import "io"
import "os"

const (
	defaultFilePerms = 644
	defaultDirPerms  = 775
)

func getFileNames(folderPath string) ([]string, error) {
	if folderPath == "" {
		folderPath = "."
	}

	folder, err := os.Open(folderPath)
	if err != nil {
		return make([]string, 0), err
	}
	defer folder.Close()

	files, err := folder.Readdirnames(-1)
	if err != nil {
		return make([]string, 0), err
	}
	return files, nil
}

func openFile(filename string) (io.WriteCloser, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaultFilePerms)
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
	return NewProcessorFromWriter(priority, w), nil
}
