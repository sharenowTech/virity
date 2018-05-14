package worker

import (
	"sync"

	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

type cGroupWorker struct {
	Group pluginregistry.ContainerGroup
	env   Environment
	F     cGroupWorkerFunc
}

type cGroupWorkerFunc struct{}

type failed struct {
	container pluginregistry.Container
	counter   int
}

// Analyse analyses scans container and sends report to monitor
func (f cGroupWorkerFunc) Analyse(env Environment, container pluginregistry.Container) error {
	stack, err := image.Analyse(container, env.CycleID, env.Scanner)
	if err != nil {
		return err
	}

	err = image.Monitor(*stack, env.CycleID, env.Monitor)
	if err != nil {
		return err
	}
	return nil
}

// Running adds container to running list but does not analyse them
func (f cGroupWorkerFunc) Running(env Environment, container pluginregistry.Container) error {
	err := env.RunningImages.Add(container)
	if err != nil {
		return err
	}
	return nil
}

// Run worker function as go func
func (cGw cGroupWorker) Run(wg *sync.WaitGroup, work ...func(env Environment, container pluginregistry.Container) error) {
	wg.Add(1)
	go func() {
		for _, F := range work {
			failChan := make(chan failed, len(cGw.Group.Container))
			for _, container := range cGw.Group.Container {
				err := F(cGw.env, container)
				if err != nil {
					log.Warn(log.Fields{
						"package":   "main/worker",
						"function":  "Run",
						"worker":    "stackWorker",
						"container": container.Name,
						"image":     container.Image,
						"hostname":  container.Hostname,
						"error":     err.Error(),
						"state":     "retry",
					}, "Operation failed - I will try again soon")
					failChan <- failed{container, 1}
					err = nil // reset error
				}

			}

			// Retry for failed operations
			for len(failChan) > 0 {
				elem := <-failChan
				log.Debug(log.Fields{
					"package":   "main/worker",
					"function":  "Run",
					"worker":    "stackWorker",
					"container": elem.container.Name,
					"image":     elem.container.Image,
					"hostname":  elem.container.Hostname,
					"retry #":   elem.counter,
					"state":     "retry",
				}, "Rerun of failed operation")
				err := F(cGw.env, elem.container)
				if err != nil {
					log.Warn(log.Fields{
						"package":   "main/worker",
						"function":  "Run",
						"worker":    "stackWorker",
						"container": elem.container.Name,
						"image":     elem.container.Image,
						"hostname":  elem.container.Hostname,
						"error":     err.Error(),
						"retry #":   elem.counter,
						"state":     "retry",
					}, "Rerun of failed operation failed again")
					if elem.counter < cGw.env.ErrorRetries {
						elem.counter++
						failChan <- failed{elem.container, elem.counter}
					} else {
						log.Error(log.Fields{
							"package":   "main/worker",
							"function":  "Run",
							"worker":    "stackWorker",
							"container": elem.container.Name,
							"image":     elem.container.Image,
							"hostname":  elem.container.Hostname,
							"error":     err.Error(),
							"retry #":   elem.counter,
							"state":     "failed",
						}, "Exceeded the maximum retries - Image will not be scanned in this cycle")
					}
					err = nil // reset error
				}

			}
		}
		wg.Done()
	}()

}
