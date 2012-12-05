/**
 * Problem:  No good way to get the average wait time of both.
**/

package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Work interface {
	Do() int64
}

type StupidWork struct {
	created time.Time
}

func (w StupidWork) Do() int64 {
	waitTime := time.Since(w.created)
	d, _ := time.ParseDuration("100ns")
	time.Sleep(d)
	return waitTime.Nanoseconds()
}

type CompressWork struct {
	created time.Time
	data    string
}

func (c *CompressWork) Do() int64 {
	waitTime := time.Since(c.created)
	var bit uint8 = 0
	d := time.Duration(1) * time.Nanosecond
	for i := 0; i < len(c.data); i++ {
		bit ^= c.data[i]
	}
	time.Sleep(d)
	return waitTime.Nanoseconds()
}

func NewWork() Work {
	stuff := "helloooooo"
	for i := 0; i < 15; i++ {
		stuff += stuff
	}
	return &CompressWork{created: time.Now(), data: stuff}
}

type Worker interface {
	Start(string, int)
	DoWork(Work)
}

type ChanWorker struct {
	MsgChan chan Work
}

func (cw *ChanWorker) Start(arg string, avgCap int) {
	if buf, err := strconv.Atoi(arg); err == nil && buf > 0 {
		cw.MsgChan = make(chan Work, buf)
	} else {
		cw.MsgChan = make(chan Work)
	}

	go func() {
		var sum int64 = 0
		var count int = 0
		for w := range cw.MsgChan {
			waitTime := w.Do()
			sum += waitTime
			count += 1
			if count >= avgCap {
				avgTime := int(sum / int64(count))
				fmt.Println(strconv.Itoa(avgTime/1000) + "us")
				sum, count = 0, 0
			}
		}
	}()
}

func (cw *ChanWorker) DoWork(w Work) {
	cw.MsgChan <- w
}

type LockWorker struct {
	Mut    sync.Mutex
	Sum    int64
	Count  int
	AvgCap int
}

func (lw *LockWorker) Start(arg string, avgCap int) {
	lw.AvgCap = avgCap
}

func (lw *LockWorker) DoWork(w Work) {
	lw.Mut.Lock()
	defer lw.Mut.Unlock()
	lw.Sum += w.Do()
	lw.Count += 1
	if lw.Count >= lw.AvgCap {
		avgTime := int(lw.Sum / int64(lw.Count))
		fmt.Println(strconv.Itoa(avgTime/1000) + "us")
		lw.Sum, lw.Count = 0, 0
	}
}

func GenerateWorkers(kind string, num int) []Worker {
	workers := make([]Worker, num)
	if kind == "chan" {
		for i := 0; i < num; i++ {
			workers[i] = &ChanWorker{}
		}
	} else if kind == "lock" {
		for i := 0; i < num; i++ {
			workers[i] = &LockWorker{}
		}
	} else {
		panic("Can't recognize kind of worker:  " + kind)
	}
	return workers
}

func StartProducers(numProds, numJobs int, workers []Worker) {
	for i := 0; i < numProds; i++ {
		go func() {
			for j := 0; j < numJobs; j++ {
				windex := j % len(workers)
				work := NewWork()
				workers[windex].DoWork(work)
			}
		}()
	}
}

func printAndExit(msg string) {
	println(msg)
	os.Exit(1)
}

func main_chanvlock() {
	if len(os.Args) == 1 {
		printAndExit(
			"Usage:\n\t" + 
			"./" + os.Args[0] + " <lock|chan> <numProducers> <numJobs> <avgCap> <numWorkers> [...workerArgs]\n\n")
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	workerType := os.Args[1]
	numProds, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printAndExit("Number of producers is not an integer:  " + os.Args[2])
	}
	numJobs, err := strconv.Atoi(os.Args[3])
	if err != nil {
		printAndExit("Number of jobs is not an integer: " + os.Args[3])
	}

	avgCap, err := strconv.Atoi(os.Args[4])
	if err != nil {
		printAndExit("The number of units to calculate the average for is not an integer.")
	}

	numWorkers, err := strconv.Atoi(os.Args[5])
	if err != nil {
		printAndExit("The number of worker resources is not an integer.")
	}

	wargs := ""
	if len(os.Args) > 6 {
		wargs = os.Args[6]
	}

	fmt.Println("Using the following params:")
	fmt.Println("\tType: " + workerType)
	fmt.Println("\tNum Prods: " + strconv.Itoa(numProds))
	fmt.Println("\tNum Jobs: " + strconv.Itoa(numJobs))
	fmt.Println("\tAvg Cap: " + strconv.Itoa(avgCap))
	fmt.Println("\tNum Workers: " + strconv.Itoa(numWorkers))
	fmt.Println("\tWorker Args: " + wargs)
	fmt.Println()

	workers := GenerateWorkers(workerType, numWorkers)
	for _, w := range workers {
		w.Start(wargs, avgCap)
	}

	StartProducers(numProds, numJobs, workers)

	time.Sleep(time.Duration(5) * time.Minute)
}
