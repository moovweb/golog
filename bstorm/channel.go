// This go file was written to test the latency on the user side when writing
// a lot of logs, or in other words, if I want to log something, what is the
// average time I'm going to wait on the call to log(...).  We want to test
// different logging implementations and see how they compare.
//
// Implementations:
//	1.	A logger contains a single channel it sends messages to where a
//			servicing routine listens on and writes to the writer.
//	2.	A logger contains a set number of channels per writer, and we have
//			multiple servicing routines.  They protect eachother with a lock.
//	3.	We avoid using channels all together and just write the log in place,
//			protected by a lock.
//
// From what we could find in our tests, it seems that locking and writing
// in place were slightly better than either channel approach, but not by much.
//
// Testing Context: 5000 routines, each writing to the log 10000 times.
//
//	Approach (1) had an average wait time of roughly 23ms per log write.
//	Approach (2) had an average wait time of roughly 23ms per log write.
//	Approach (3) had an average wait time of roughly 21ms per log write.
//
// Other Findings:
//
//	runtime.GOMAXPROCS(runtime.NumCPU()) had a huge improvement in performance.
//	Nearly halfed the wait time.
//
package main

import (
	"fmt"
	"golog"
	"net"
	"strconv"
	"sync"
	"time"

//	"runtime"
)

const bufSize = 512

var mu *sync.Mutex = &sync.Mutex{}

func lock_service(msgChan <-chan string, syslog net.Conn) {
	fmt.Println("Locked Service started.")
	for s := range msgChan {
		mu.Lock()
		syslog.Write([]byte(s))
		mu.Unlock()
	}
}

func service(msgChan <-chan string, syslog net.Conn) {
	fmt.Println("Service started.")
	for s := range msgChan {
		syslog.Write([]byte(s))
	}
}

func clog(msgChan chan<- string, reps int) {
	t := time.Now()
	for i := 0; i < reps; i++ {
		msgChan <- "hi"
	}
	sum := int64(time.Since(t))
	avg := int(sum / int64(reps))
	fmt.Println("Average wait time:  " + strconv.Itoa(avg))
}

func clogWrite(syslog net.Conn, reps int) {
	t := time.Now()
	for i := 0; i < reps; i++ {
		mu.Lock()
		syslog.Write([]byte("hi"))
		mu.Unlock()
	}
	sum := int64(time.Since(t))
	avg := int(sum / int64(reps))
	fmt.Println("Average wait time:  " + strconv.Itoa(avg))
}

func test_single_chan(syslog net.Conn, num_routines, num_writes int) {
	msgChan := make(chan string, bufSize)

	go service(msgChan, syslog)

	reps := num_routines
	for i := 0; i < reps; i++ {
		go clog(msgChan, num_writes)
	}
	d, _ := time.ParseDuration("1000s")
	time.Sleep(d)
	close(msgChan)
}

func test_multi_chan(syslog net.Conn, num_routines, num_writes, num_chans int) {
	chans := make([]chan string, num_chans)
	for i := 0; i < num_chans; i++ {
		chans[i] = make(chan string, bufSize)
		go lock_service(chans[i], syslog)
	}

	reps := num_routines
	for i := 0; i < reps; i++ {
		go clog(chans[i%num_chans], num_writes)
	}

	d, _ := time.ParseDuration("1000s")
	time.Sleep(d)
	for i := 0; i < num_chans; i++ {
		close(chans[i])
	}
}

func test_nochan_lock(syslog net.Conn, num_routines, num_writes int) {
	reps := num_routines
	for i := 0; i < reps; i++ {
		go clogWrite(syslog, num_writes)
	}

	d, _ := time.ParseDuration("1000s")
	time.Sleep(d)
}

func main() {
	//	runtime.GOMAXPROCS(runtime.NumCPU())
	syslog, err := golog.DialSyslog("", "")
	if err != nil {
		fmt.Println("Couldn't coonect to syslog:  " + err.Error())
	}

	//	test_single_chan(syslog, 5000, 10000)
	//	test_multi_chan(syslog, 5000, 10000, 10)
	// test_nochan_lock(syslog, 5000, 10000)
	_ = syslog

}
