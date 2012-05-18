// This go file was written as a test to see how much latency using read/write
// locks adds to a simple variable read.  Running this program with and without
// the use of read/write locks showed that the rw.RLock()/RUnlock() methods add
// roughly 20ns of time to a read.
//
// This test was performed in order to judge the way we would implement the
// ability to change priorities of processors on the fly.  We considered the
// following implementations:
//
// 1)	Wrapping the priority field in the Processor with a RW lock and offering
//		Setter/Getter methods for it.
// 2)	Pushing the checking/filtering of the priority of a message to the 
//		writer go routine which writes the message to the appropriate writer
// 3)	Each go routine in the main application has its own logger and thus,
//		will have to update the priority accordingly.  However, since the logger
//		is only used in a single go routine, there's no race conditions in 
//		simply updating the priority.
//
// We chose solution (1) after the test below since the penalty for adding
// the RW locks is small and offers more flexibilty in terms of how a logger
// can be used (aka, we can have one global logger for all go routines, etc...)
//
// Solution (2) is elegant in the sense that it would require no extra locking,
// however, it would mean that the application as a whole would have to do more
// work as messages that would be filtered out before hand will now have to 
// go through the log channel (then to get filtered out).
//
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	//teapot := 8675309
	numchan := make(chan int, 100)
	go func() {
		for {
			numchan <- 5
		}
	}()
	//	rw := &sync.RWMutex{}
	finisher := &sync.WaitGroup{}
	//starter.Add(1)
	for i := 0; i < 400; i++ {
		finisher.Add(1)
		go func() {
			j := 0
			//starter.Wait()
			now := time.Now()
			for ; j < 10000000; j++ {
				<-numchan
			}
			duration := time.Now().Sub(now)
			avg := float64(duration.Nanoseconds()) / float64(j)
			fmt.Printf("Avg: %fns\n", avg)
			finisher.Done()
		}()
	}
	finisher.Wait()
}
