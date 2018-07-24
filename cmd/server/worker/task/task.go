package task

import (
	"sync"

	"github.com/car2go/virity/cmd/server/image/model"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Queue is the task queue for the workers
var Queue chan Task

// Monitored (global map to persist currently monitored images)
var monitored = model.Model{}

// BaseTask is a template for specific tasks. Every Task has a base task. All subtasks share the same waitgroup and RunningImages list.
type BaseTask struct {
	RunningImages *model.Model
	Store         pluginregistry.Store
	Scanner       pluginregistry.Scan
	Monitor       pluginregistry.Monitor
	Retries       int
	CycleID       int
	wg            *sync.WaitGroup
}

// Task specifies all functions, a task has to implement
type Task interface {
	// Run is called when a task is executed
	Run() error
	// Retry is called to add a failed task again to the queue to retry
	Retry()
	// Adds the task to the WaitGroup
	Register()
	// Sets the task to Done in the WaitGroup
	DeRegister()
}

func init() {
	Queue = make(chan Task)
}

// AddToQueue adds a new Task to the TaskQueue and adds it to the WaitGroup
func AddToQueue(t Task) {
	t.Register()
	Queue <- t
}

// New creates a new Basetask.
func New(wg *sync.WaitGroup, cycleID int, maxRetries int, store pluginregistry.Store, scanner pluginregistry.Scan, monitor pluginregistry.Monitor) BaseTask {
	return BaseTask{
		RunningImages: &model.Model{},
		wg:            wg,
		Store:         store,
		Scanner:       scanner,
		Monitor:       monitor,
		CycleID:       cycleID,
		Retries:       maxRetries,
	}
}

// Container creates a Subtask of a Basetask to handle containers
func (b BaseTask) Container(data pluginregistry.Container, f ...func(container) error) container {
	return container{
		Base:      b,
		Container: data,
		Process:   f,
	}
}

// Maintain creates a Subtask of a Basetask to maintain the environment (data store etc.)
func (b BaseTask) Maintain(f ...func(maintain) error) maintain {
	return maintain{
		Base:    b,
		Process: f,
	}
}
