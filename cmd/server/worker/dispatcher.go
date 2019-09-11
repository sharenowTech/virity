package worker

import (
	"github.com/sharenowTech/virity/cmd/server/worker/task"
	"github.com/sharenowTech/virity/internal/log"
)

// Dispatcher holds workers and the WorkerPool to assign tasks
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan task.Task
	maxWorkers int
}

// NewDispatcher creates new Dispatcher
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan task.Task, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers}
}

// Run the Dispatcher
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.WorkerPool)
		worker.Start()
		log.Debug(log.Fields{
			"package":  "worker",
			"function": "Dispatcher/Run",
		}, "Worker created")
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case t := <-task.Queue:
			log.Debug(log.Fields{
				"package":  "worker",
				"function": "Dispatcher/dispatch",
			}, "Task request received")
			// a task request has been received
			go func(task task.Task) {
				// try to obtain a worker task channel that is available.
				// this will block until a worker is idle
				taskChannel := <-d.WorkerPool

				// dispatch the task to the worker task channel
				taskChannel <- task
			}(t)
		}
	}
}
