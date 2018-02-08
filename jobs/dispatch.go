package jobs

type Dispatcher struct {
	JobQueue   chan Job
	WorkerPool chan chan Job
	maxWorkers int
}

func NewDispatcher(jq chan Job, maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{JobQueue: jq, WorkerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		// Get a job from the provided queue
		case job := <-d.JobQueue:
			go func(j Job) {
				// Will block until a worker is idle
				jobChan := <-d.WorkerPool
				// Give that worker the job
				jobChan <- job
			}(job)
		}
	}
}
