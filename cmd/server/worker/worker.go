package worker

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Worker represents the worker that executes the job
type worker struct {
	WorkerPool  chan chan Task
	TaskChannel chan Task
	quit        chan bool
}

func newWorker(workerPool chan chan Task) worker {
	return worker{
		WorkerPool:  workerPool,
		TaskChannel: make(chan Task),
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
			case task := <-w.TaskChannel:
				// we have received a work request.
				task.Work()

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

// Init creates a new Environment and initializes everything
func Init(cycle int, store pluginregistry.Store, scanner pluginregistry.Scan, monitor pluginregistry.Monitor) Environment {
	return Environment{
		RunningImages: &image.Active{},
		Store:         store,
		Scanner:       scanner,
		Monitor:       monitor,
		ErrorRetries:  5,
		CycleID:       cycle,
	}
}
