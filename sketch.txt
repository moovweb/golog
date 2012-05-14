import "io"

type LogEntry struct {
	w io.Writer
	msg string
}

type LowlevelWriter struct {
	io.Writer //Syslog //file
	LogChann chan *LogEntry
}

(lw *LogWriter) WriteString(data string) error {
	entry := LogEntry{w: lw.Writer, msg: data}
	LogChann <- entry
}

(s *SyslogWriter) WriteString(data string) error {
	rawData := fmt.Fprintf(sock, "<%d>%s %s %s: %s\n", s.facility * 8+int(s.level), timestr, host, s.Prefix, data)
}

NewSyslogWriter() {

}
NewFileLogWriter()
NewConsoleLogWriter()

type SyslogWriter struct {
	*LowlevelWriter
	level int
	facility int
	prefix string
}


type LogWriter interace {
	Logf(level int, format string, args interface{}...)
}


(lw *LogWriter) Logf(level int, format string, args interface{}...) {
	if level >= lw.Level {
	data = fmt.Sprintf(format, args)
	lw.WriteString(data)
}
}

type Logger struct {
	writers []*LogWriter
}

(l *Logger) Info(format string, args interface{}...) {

}

(l *Logger) Log(level int, format string, args interface{}...) {

}







