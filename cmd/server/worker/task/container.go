package task

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// container is a task for a container
type container struct {
	Base      BaseTask
	Container pluginregistry.Container
	Process   []func(container) error
}

// Analyse analyses scans container and sends report to monitor
func Analyse(c container) error {
	log.Debug(log.Fields{
		"package":  "worker",
		"function": "Analyse",
	}, "Run task")
	stack, err := image.Analyse(c.Container, c.Base.CycleID, c.Base.Scanner)
	if err != nil {
		return err
	}

	err = image.Monitor(*stack, c.Base.CycleID, c.Base.Monitor)
	if err != nil {
		return err
	}
	return nil
}

// Running adds container to running list but does not analyse them
func Running(c container) error {
	log.Debug(log.Fields{
		"package":  "worker",
		"function": "Running",
	}, "Run task")
	err := c.Base.RunningImages.Add(c.Container)
	if err != nil {
		return err
	}
	return nil
}

// Run worker function as go func
func (c container) Run() error {
	for _, proc := range c.Process {
		err := proc(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Retry adds the task to the queue until the counter is 0
func (c *container) Retry() {
	if c.Base.Retries <= 0 {
		log.Warn(log.Fields{
			"package":  "worker",
			"function": "Container/Retry",
		}, "Task failed after several retries. It will be dropped")
		c.Base.wg.Done()
		return
	}
	log.Info(log.Fields{
		"package":  "worker",
		"function": "Container/Retry",
	}, "Task failed. I will retry.")
	c.Base.Retries--
	Queue <- c
	return
}

// Register adds Task to a WaitGroup
func (c *container) Register() {
	c.Base.wg.Add(1)
}

// DeRegister sets the Task in the WaitGroup to Done
func (c *container) DeRegister() {
	c.Base.wg.Done()
}
