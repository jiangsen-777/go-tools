package main

import (
	"sync"
)

type Job func()

type Worker struct {
	ID         int
	WorkerPool chan chan Job   // 协程池的工作队列
	JobChannel chan Job        // job队列，每次可以放一个任务
	quit       chan bool       // 停止信号
	wg         *sync.WaitGroup // 协程等待
}

type Pool struct {
	JobQueue            chan Job
	WorkerPool          chan chan Job
	MaxWorkers          int
	MaxQueue            int
	WorkerArray         []*Worker
	quit                chan bool
	wg                  *sync.WaitGroup
	CurrentRunningCount int
}

func NewWorker(id int, workerPool chan chan Job, wg *sync.WaitGroup) Worker {
	return Worker{
		ID:         id,
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		wg:         wg,
	}
}

func (w Worker) Start() {
	go func() {
		w.wg.Add(1)
		for {
			w.WorkerPool <- w.JobChannel // 完成任务后将本协程的任务队列放入协程池的工作队列
			select {
			case job := <-w.JobChannel: // 从本协程的任务队列拿取任务运行
				job()
			case <-w.quit:
				w.wg.Done()
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func NewPool(maxWorkers, maxQueue int) *Pool {
	workerPool := make(chan chan Job, maxWorkers)
	jobQueue := make(chan Job, maxQueue)
	workerArray := make([]*Worker, maxWorkers)
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		worker := NewWorker(i, workerPool, &wg)
		workerArray[i] = &worker
	}

	return &Pool{
		JobQueue:    jobQueue,
		WorkerPool:  workerPool,
		MaxWorkers:  maxWorkers,
		MaxQueue:    maxQueue,
		WorkerArray: workerArray,
		quit:        make(chan bool),
		wg:          &wg,
	}
}

func (p *Pool) Start() {
	for _, worker := range p.WorkerArray {
		worker.Start()
	}
	go p.dispatch()
}

func (p *Pool) Stop() {
	go func() {
		p.quit <- true
	}()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) dispatch() {
	for {
		select {
		case job := <-p.JobQueue: // 从主job队列取出job
			workerChannel := <-p.WorkerPool // 从工作队列取出可用的协程的job队列
			workerChannel <- job            // 将job放入协程的job队列
		case <-p.quit:
			for _, worker := range p.WorkerArray {
				worker.Stop()
			}
			return
		}
	}
}

/*
func main() {
	pool := NewPool(3, 10)
	pool.Start()
	for i := 0; i < 20; i++ {
		count := i
		pool.JobQueue <- func() {
			fmt.Printf("Job %d started\n", count)
			time.Sleep(1 * time.Second)
			fmt.Printf("Job %d finished\n", count)
		}
	}
	pool.Stop()
	pool.Wait()
}
*/
