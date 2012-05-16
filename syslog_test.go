package golog

import (
	"testing"
	"net"
	"time"
	"strings"
	"strconv"
)

// When reading from the socket, we read null chars if nothing is left to 
// read, so we have to trim them out before we do our comparison.
func trimNull(s string) string {
	if s == "" {
		return s
	}
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] != 0x00 {
			break
		}
	}
	var j int
	for j = len(s)-1; j >= 0; j-- {
		if s[j] != 0x00 {
			break
		}
	}
	return s[i:j]
}

func checkOutput(result string, f Facility, p Priority, prefix, msg string, t *testing.T) {
	expectedStart := "<" + strconv.Itoa(int(f) * 8 + int(p)) + ">"
	expectedEnd := p.String() + ": " + prefix + strings.TrimSpace(msg)
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
		if err != nil || n == 0 {
			break
		}
		rcvd += string(buf[0:])
	}
	msgChan <- trimNull(rcvd)
	c.Close()
}

func startTimedServer(msgChan chan<- string, ms int) (serverAddr string, err error){
	c, e := net.ListenPacket("udp", "127.0.0.1:0")
	if e != nil {
		return "", e
	}
	serverAddr = c.LocalAddr().String()
	c.SetReadDeadline(time.Now().Add(time.Duration(ms) * time.Millisecond))
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

func TestSingleLogWrite(t *testing.T) {
	msgChan := make(chan string)
	servAddy, err := startTimedServer(msgChan, 100)
	if err != nil {
		t.Errorf("Couldn't start syslog listener:  %s", err.Error())
	}

	prefix := "syslog_test: "
	facility := LOCAL0
	minPriority := LOG_DEBUG
	message := "Testing Info."
	priority := LOG_INFO

	logger := createSyslogger(servAddy, prefix, facility, minPriority, t)

	logger.Info(message)
	rcvd := <-msgChan
	checkOutput(rcvd, facility, priority, prefix, message + "\n", t)
	closeSyslog(logger)
}

func TestConcurrentSyslogWrite(t *testing.T) {
	
}
