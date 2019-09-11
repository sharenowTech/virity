package worker

import (
	"github.com/sharenowTech/virity/cmd/server/worker/task"
	"github.com/sharenowTech/virity/internal/log"
)

// Worker represents the worker that executes the job
type worker struct {
	WorkerPool  chan chan task.Task
	TaskChannel chan task.Task
	quit        chan bool
}

// newWorker creates a new worker
func newWorker(workerPool chan chan task.Task) worker {
	return worker{
		WorkerPool:  workerPool,
		TaskChannel: make(chan task.Task),
		quit:        make(chan bool),
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.TaskChannel

			select {
			case t := <-w.TaskChannel:
				// we have received a work request.
				err := t.Run()
				if err != nil {
					log.Debug(log.Fields{
						"package":  "worker",
						"function": "Start",
						"Error":    err.Error(),
					}, "Task failed. I will ask for a retry.")
					t.Retry()
					continue
				}
				t.DeRegister()

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
