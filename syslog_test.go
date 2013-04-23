package golog

import (
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func checkOutput(result string, f Facility, p Priority, prefix, msg string, t *testing.T) {
	expectedStart := "<" + strconv.Itoa(int(f)*8+int(p)) + ">"
	expectedEnd := p.String() + ": " + prefix + msg
	if !strings.HasPrefix(result, expectedStart) || !strings.HasSuffix(result, expectedEnd) {
		errmsg := "Failed log consistency check:\nExpected '%s'\nResult   '%s'"
		t.Errorf(errmsg, expectedStart+" ... "+expectedEnd, result)
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
		rcvd += string(buf[0:n])
	}

	msgChan <- rcvd
	c.Close()
}

func startServer(msgChan chan<- string) (serverAddr string, err error) {
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
	if runtime.GOOS == "windows" {
		return
	}
	checkSyslogNewProcessor(LOCAL0, LOG_DEBUG, t)
}

func TestDialSyslog(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
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
	logger.Close()
}

func checkSyslogPost(f Facility, p Priority, t *testing.T) {
	msgChan := make(chan string)
	servAddy, err := startServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start syslog listener:  %s", err.Error())
	}

	prefix := "syslog_single_test: "
	minPriority := LOG_DEBUG
	message := "Testing."

	logger := createSyslogger(servAddy, prefix, f, minPriority, t)

	logger.Logf(p, message)
	rcvd := <-msgChan
	checkOutput(rcvd, f, p, prefix, message+"\n", t)

	closeSyslog(logger)
}

func TestSingleLogWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	checkSyslogPost(LOCAL0, LOG_INFO, t)
}

func TestMultipleLogWrites(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	for _, f := range SyslogFacilities() {
		for _, p := range Priorities() {
			checkSyslogPost(f, p, t)
		}
	}
}

func TestConcurrentSyslogWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	msgChan := make(chan string)
	servAddy, err := startServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start syslog listener:  %s", err.Error())
	}

	prefix := "syslog_conc_test: "
	facility := LOCAL0
	minPriority := LOG_DEBUG

	logger := createSyslogger(servAddy, prefix, facility, minPriority, t)

	total_routines := 50
	var wg sync.WaitGroup
	for i := 0; i < total_routines; i++ {
		var tmp int = i
		wg.Add(1)
		go func() {
			logger.Infof("Testing routine %08d", tmp)
			wg.Done()
		}()
	}
	wg.Wait()

	dur, _ := time.ParseDuration("2s")
	time.Sleep(dur)

	logs := <-msgChan
	log_lines := strings.Split(strings.TrimSpace(logs), "\n")
	sort.Strings(log_lines)
	if len(log_lines) != total_routines {
		errmsg := "Some log lines are missing! Expected %d, Found %d"
		t.Fatalf(errmsg, total_routines, len(log_lines))
	}
}
