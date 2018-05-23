package task

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Container is a task for a container
type Container struct {
	Base      BaseTask
	Container pluginregistry.Container
	Process   []func(Container) error
}

// Analyse analyses scans container and sends report to monitor
func Analyse(c Container) error {
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
func Running(c Container) error {
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
func (c Container) Run() error {
	for _, proc := range c.Process {
		err := proc(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) Retry() {
	if c.Base.Retries > 0 {
		return
	}
	c.Base.Retries--
	AddToQueue(c)
	return
}

func (c *Container) Register() {
	c.Base.wg.Add(1)
}

func (c *Container) DeRegister() {
	c.Base.wg.Done()
}
