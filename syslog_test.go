package golog

import (
	"testing"
)

func checkSyslogNewProcessor(f Facility, p Priority, t *testing.T) {
	proc, err := NewSyslogProcessor(f, p)
	if err != nil {
		t.Fatalf("NewSyslogProcessor(f=%d, p=%d) failed:  %s", f, p, err.Error())
	} else {
		proc.Close()
	}
}

func TestNew(t *testing.T) {
	checkSyslogNewProcessor(LOCAL0, LOG_DEBUG, t)
	// We could check all combinations here, but as of right now,
	// creating a new syslog processor does not depend on facility or priority.
	// If it ever does, then we should add the rest of the combinations.
}

func TestDialSyslog(t *testing.T) {
	conn, err := DialSyslog("", "")
	if err != nil {
		t.Fatalf("Couldn't connect to syslog:  %s", err.Error())
	} else {
		conn.Close()
	}
}
