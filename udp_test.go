package golog

import "testing"
import "net"
import "time"
import "sync"
import "strings"
import "sort"

func checkUdpOutput(result string, p Priority, prefix, msg string, t *testing.T) {
	expected := msg
	if result != expected {
		errmsg := "Failed log consistency check:\nExpected '%s'\nResult   '%s'"
		t.Errorf(errmsg, expected, result)
	}
}

func runUdpReader(c net.PacketConn, msgChan chan<- string) {
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

func startUdpServer(msgChan chan<- string) (host string, err error) {
	c, e := net.ListenPacket("udp", "localhost:0")
	if e != nil {
		return "", e
	}
	host = c.LocalAddr().String()
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	go runUdpReader(c, msgChan)
	return host, nil
}

func checkUdpNewProcessor(host string, p Priority, t *testing.T) {
	proc, err := NewUdpProcessorAt(host, p)
	if err != nil {
		t.Fatalf("NewUdpProcessor(host=%s, p=%d) failed:  %s", host, p, err.Error())
	} else {
		proc.Close()
	}
}

func TestNewUdpProcessor(t *testing.T) {
	checkUdpNewProcessor("localhost:8675", LOG_DEBUG, t)
}

func TestDialUdp(t *testing.T) {
	conn, err := DialUdp("localhost:8675")
	if err != nil {
		t.Fatalf("Couldn't connect to udp:  %s", err.Error())
	} else {
		conn.Close()
	}
}

func createUdpLogger(host string, prefix string, p Priority, t *testing.T) *Logger {
	udpProc, err := NewUdpProcessorAt(host, p)
	if err != nil {
		t.Fatalf("Coudln't create processor to listen to udp: %s", err.Error())
	}
	logger := NewLogger(prefix)
	logger.AddProcessor("udp", udpProc)
	return logger
}

// essentially closes the log
func closeUdpLogger(logger *Logger) {
	logger.Close()
}

func checkUdpPost(p Priority, t *testing.T) {
	msgChan := make(chan string)
	host, err := startServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start udp listener:  %s", err.Error())
	}

	prefix := "udp_single_test: "
	minPriority := LOG_DEBUG
	message := "Testing."

	logger := createUdpLogger(host, prefix, minPriority, t)

	logger.Logf(p, message)
	rcvd := <-msgChan
	checkUdpOutput(rcvd, p, prefix, message+"\n", t)

	closeUdpLogger(logger)
}

func TestUdpSingleLogWrite(t *testing.T) {
	checkUdpPost(LOG_INFO, t)
}

func TestUdpMultipleLogWrites(t *testing.T) {
	for _, p := range Priorities() {
		checkUdpPost(p, t)
	}
}

func TestConcurrentUdpWrite(t *testing.T) {
	msgChan := make(chan string)
	host, err := startUdpServer(msgChan)
	if err != nil {
		t.Fatalf("Couldn't start udp listener:  %s", err.Error())
	}

	prefix := "udp_conc_test: "
	minPriority := LOG_DEBUG

	logger := createUdpLogger(host, prefix, minPriority, t)

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
