package jobs

import (
	"net"
	"net/rpc"
)

const (
	MAX_WORKER = 1000
	MAXQUEUE   = 1000
)

type Job interface {
	Execute()
}

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			// Put this workers channel back into the pool - "I'm not doing anything"
			w.WorkerPool <- w.JobChannel
			select {
			// The channel has been taken off the worker pool and provided a job
			case job := <-w.JobChannel:
				job.Execute()

			// Worker has been told to stop
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type TransactionJob struct {
	Conn net.Conn
}

func (j TransactionJob) Execute() {
	rpc.ServeConn(j.Conn)
}
