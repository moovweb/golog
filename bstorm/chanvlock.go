/**
 * Problem:  No good way to get the average wait time of both.
**/

package main

import (
	"time"
	"strconv"
	"sync"
	"runtime"
	"os"
)


type Work interface {
	Do() int64
}

type StupidWork struct {
	created time.Time
	data string
}

func (w StupidWork) Do() int64 {
	waitTime := time.Since(w.created)
	d, _ := time.ParseDuration("100ns")
	time.Sleep(d)
	return waitTime.Nanoseconds()
}

func NewWork() Work {
	return &StupidWork { created: time.Now(), data: "Hey, listen..." }
}

type Worker interface {
	Start(string, int)
	DoWork(Work)
}

type ChanWorker struct {
	MsgChan chan Work
}

func (cw *ChanWorker) Start(arg string, avgCap int) {
	if buf, err := strconv.Atoi(arg); err == nil && buf > 0{
		cw.MsgChan = make(chan Work, buf)
	} else {
		cw.MsgChan = make(chan Work)
	}

	go func() {
		var sum int64 = 0
		var count int = 0
		for w := range(cw.MsgChan) {
			waitTime := w.Do()
			sum += waitTime
			count += 1
			if count >= avgCap {
				avgTime := int(sum/int64(count))
				println(strconv.Itoa(avgTime/1000) + "us")
				sum, count = 0, 0
			}
		}
	}()
}

func (cw *ChanWorker) DoWork(w Work) {
	cw.MsgChan <- w
}


type LockWorker struct {
	Mut sync.Mutex
	Sum int64
	Count int
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
		avgTime := int(lw.Sum/int64(lw.Count))
		println(strconv.Itoa(avgTime/1000) + "us")
		lw.Sum, lw.Count = 0, 0
	}
}


func GenerateWorkers(kind string, num int) ([]Worker) {
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



var numProds, numJobs, avgCap int = 5000, 10000, 50000
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	workerType := os.Args[1]
	numWorkers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		numWorkers = 5
	}
	wargs := ""
	if len(os.Args) > 3 {
		wargs = os.Args[3]
	}
	

	workers := GenerateWorkers(workerType, numWorkers)
	for _, w := range(workers) {
		w.Start(wargs, avgCap)
	}

	StartProducers(numProds, numJobs, workers)

	time.Sleep(time.Duration(5) * time.Minute)
}
