package task

import (
	"sync"

	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Queue represents the task queue for the workers
var Queue chan Task

type Func struct{}

type BaseTask struct {
	RunningImages *image.Active
	Store         pluginregistry.Store
	Scanner       pluginregistry.Scan
	Monitor       pluginregistry.Monitor
	Retries       int
	CycleID       int
	wg            *sync.WaitGroup
}

type Task interface {
	Run() error
	Retry()
	Register()
	DeRegister()
}

func init() {
	Queue = make(chan Task)
}

func AddToQueue(t Task) {
	Queue <- t
}

func NewContainer(base BaseTask, wg *sync.WaitGroup, container pluginregistry.Container, f ...func(Container) error) Container {
	base.wg = wg
	return Container{
		Base:      base,
		Container: container,
		Process:   f,
	}
}

func NewMaintain(base BaseTask, wg *sync.WaitGroup, f ...func(Maintain) error) Maintain {
	base.wg = wg
	return Maintain{
		Base:    base,
		Process: f,
	}
}
