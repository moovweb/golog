package golog

import (
	"net"
	"os"
	"fmt"
)

type Priority int

const (
	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

type Facility int

const (
	LOCAL0 = iota + 16
	LOCAL1
	LOCAL2
	LOCAL3
	LOCAL4
	LOCAL5
	LOCAL6
	LOCAL7
)

type SyslogWriter struct {
	conn net.Conn
}

func connectToSyslog() (sock net.Conn, err error) {
	logTypes := []string { "unixgram", "unix" }
	logPaths := []string { "/dev/log", "/var/run/syslog" }

	for _, network := range logTypes {
		for _, path := range logPaths {
			sock, err = net.Dial(network, path)
			if err == nil {
				fmt.Fprintf(os.Stderr, "syslog uses %s:%s\n", network, path)
				return sock, nil
			}
		}
	}
	return nil, err
	
}

func NewSyslogWriter() (w *SyslogWriter, err error) {
	sock, err := connectToSyslog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error SyslogWriter: %s\n", err.Error())
		return nil, err
	}

	return &SyslogWriter { conn: sock }, nil
}

func (w *SyslogWriter) Write(b []byte) (int, error) {
	return w.conn.Write(b)
}

