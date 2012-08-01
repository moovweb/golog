package golog

import "io"
import "net"

type UdpWriter struct {
	noConn io.WriteCloser
}

func (nw *UdpWriter) Close() error {
	return nw.noConn.Close()
}

func (nw *UdpWriter) Write(data []byte) (n int, err error) {
	return nw.noConn.Write(data)
}

func DialUdp(host string) (sock io.WriteCloser, err error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	return &UdpWriter{noConn: conn}, nil
}

type UdpProcessor struct {
	*DefaultProcessor
}

func (np *UdpProcessor) Process(entry *LogEntry) {
	if entry.Priority <= np.GetPriority() {
		msg := entry.Msg
		np.Dispatcher.Send(msg)
	}
}

func NewUdpProcessorAt(host string, p Priority) (LogProcessor, error) {
	dw, err := DialUdp(host)
	if err != nil {
		return nil, err
	}
	dsp := NewLogDispatcher(dw)
	defaultProcessor := NewProcessor(p, dsp, true).(*DefaultProcessor)
	return &UdpProcessor{DefaultProcessor: defaultProcessor}, nil
}

func NewUdpProcessor(p Priority) (LogProcessor, error) {
	return NewUdpProcessorAt("localhost:0", p)
}
