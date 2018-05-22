package worker

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Queue represents the task queue for the workers
var Queue chan Task

type Task interface {
	Work()
}

type Environment struct {
	RunningImages *image.Active
	Store         pluginregistry.Store
	Scanner       pluginregistry.Scan
	Monitor       pluginregistry.Monitor
	ErrorRetries  int
	CycleID       int
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Task
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Task, maxWorkers)
	return &Dispatcher{WorkerPool: pool}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case task := <-Queue:
			// a task request has been received
			go func(task Task) {
				// try to obtain a worker task channel that is available.
				// this will block until a worker is idle
				taskChannel := <-d.WorkerPool

				// dispatch the task to the worker task channel
				taskChannel <- task
			}(task)
		}
	}
}
