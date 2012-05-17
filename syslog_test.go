package golog

import (
	"testing"
	"time"
	"net"
	"strings"
	"strconv"
	"sync"
	"sort"
)

func checkOutput(result string, f Facility, p Priority, prefix, msg string, t *testing.T) {
	expectedStart := "<" + strconv.Itoa(int(f) * 8 + int(p)) + ">"
	expectedEnd := p.String() + ": " + prefix + msg
	if !strings.HasPrefix(result, expectedStart) || !strings.HasSuffix(result, expectedEnd) {
		errmsg := "Failed log consistency check:\nExpected '%s'\nResult   '%s'"
		t.Errorf(errmsg, expectedStart + " ... " + expectedEnd, result)
	}
}

func runSyslogReader(c net.PacketConn, msgChan chan<- string) {
	var buf [4096]byte
	var rcvd string = ""
	for {
		n, _, err := c.ReadFrom(buf[0:])
		if err != nil || n == 0 { break }
		rcvd += string(buf[0:n])
	}

	msgChan <- rcvd
	c.Close()
}

func startServer(msgChan chan<- string) (serverAddr string, err error){
	c, e := net.ListenPacket("udp", "127.0.0.1:0")
	if e != nil {
		return "", e
	}
	serverAddr = c.LocalAddr().String()
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	go runSyslogReader(c, msgChan)
	return serverAddr, nil
}

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

func createSyslogger(servAddy, prefix string, f Facility, p Priority, t *testing.T) *Logger {
	sysProc, err := NewSyslogProcessorAt("udp", servAddy, f, p)
	if err != nil {
		t.Fatalf("Coudln't create processor to listen to syslog: %s", err.Error())
	}
	logger := NewLogger(prefix)
	logger.AddProcessor("syslog", sysProc)
	return logger
}

// essentially closes the log
func closeSyslog(logger *Logger) {
	logger.AddProcessor("syslog", nil)
}

func checkSyslogPost(f Facility, p Priority, t *testing.T) {
	msgChan := make(chan string)
	servAddy, err := startServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start syslog listener:  %s", err.Error())
	}

	prefix := "syslog_single_test: "
	minPriority := LOG_DEBUG
	message := "Testing Info."

	logger := createSyslogger(servAddy, prefix, f, minPriority, t)

	logger.Log(p, message)

	rcvd := <-msgChan
	checkOutput(rcvd, f, p, prefix, message + "\n", t)
	closeSyslog(logger)
}

func TestSingleLogWrite(t *testing.T) {
	checkSyslogPost(LOCAL0, LOG_INFO, t)
}

func TestMultipleLogWrites(t *testing.T) {
	for p := range(Priorities()) {
		for f := range(SyslogFacilities()) {
			checkSyslogPost(Facility(f), Priority(p), t)
		}
	}
}

func TestConcurrentSyslogWrite(t *testing.T) {
	msgChan := make(chan string)
	servAddy, err := startServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start syslog listener:  %s", err.Error())
	}

	prefix := "syslog_conc_test: "
	facility := LOCAL0
	minPriority := LOG_DEBUG
	
	logger := createSyslogger(servAddy, prefix, facility, minPriority, t)

	total_routines := 5000
	var wg sync.WaitGroup
	for i := 0; i < total_routines; i++ {
		var tmp int = i
		wg.Add(1)
		go func() {
			logger.Info("Testing routine %08d", tmp)
			wg.Done()
			println ("finished: " + strconv.Itoa(tmp))
		}()
	}
	wg.Wait()
	
	logs := <-msgChan
	log_lines := strings.Split(strings.TrimSpace(logs), "\n")
	sort.Strings(log_lines)
	if len(log_lines) != total_routines {
		errmsg := "Some log lines are missing! Expected %d, Found %d"
		t.Fatalf(errmsg, total_routines, len(log_lines))
	}
}



