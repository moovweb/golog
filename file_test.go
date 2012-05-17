package golog

import (
	"testing"
	"io/ioutil"
	"strings"
	"time"
	"sync"
)


const file_prefix string = "golog_filetest"

func readLogFile(filename string) (loglines []string, err error) {
	file_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}

	log := string(file_data)
	log = strings.TrimSpace(log)
	return strings.Split(log, "\n"), nil
}

func TestDifferentPriorities(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", file_prefix)
	if err != nil {
		t.Fatalf("Couldn't open tmp file: %s", err.Error())
	}

	proc := NewProcessorFromWriter(LOG_DEBUG, tmpfile)
	logger := NewLogger("file_test: ")
	logger.AddProcessor("file", proc)
	
	for _, p := range(Priorities()) {
		logger.Log(p, "Hey, listen...")
	}

	d, _ := time.ParseDuration("100ms")
	time.Sleep(d)
	loglines, err := readLogFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file:  %s", err.Error())
	}

	if len(loglines) != len(Priorities()) {
		msg := "Unexpected number of log lines ouputed.  Expected %d, but was %d"
		t.Fatalf(msg, len(Priorities()), len(loglines))
	}

	for i, p := range(Priorities()) {
		expected := p.String() + ": file_test: Hey, listen..."
		if loglines[i] != expected {
			t.Errorf("Unexpected log line.\nExpected: %s\nBut was: %s", expected, loglines[i])
		}
	}
}

func TestConcurrentLogging(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "golog_conc_file_test")
	if err != nil {
		t.Fatalf("Couldn't create tmp file: %s", err.Error())
	}
	proc := NewProcessorFromWriter(LOG_DEBUG, tmpfile)
	logger := NewLogger("conc_test: ")
	logger.AddProcessor("file", proc)

	wg := &sync.WaitGroup{}
	num_routines := 5000
	for i := 0; i < num_routines; i++ {
		wg.Add(1)
		go func () {
			logger.Notice("Help! I need somebody!")
			logger.Warning("Help! Not just anybody!")
			logger.Error("Help! You know I need someone!")
			logger.Critical("Heeeeeeeeeeeeelp!")
			wg.Done()
		}()
	}
	wg.Wait()
	d, _ := time.ParseDuration("100ms")
	time.Sleep(d)
	
	loglines, err := readLogFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error when reading tmp file:  %s", err.Error())
	}

	if len(loglines) != (num_routines * 4) {
		msg := "Unexpected number of log lines.  Expected %d, but was %d"
		t.Fatalf(msg, num_routines * 4, len(loglines))
	}
}
