package worker

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
)

type maintainWorker struct {
	env Environment
	F   maintainWorkerFunc
}

type MaintainTask struct {
	Env  Environment
	Func func(env Environment) error
}

type maintainWorkerFunc struct{}

// Resolve resolves monitoring alerts for not running images and resets the list of running images
func (f maintainWorkerFunc) Resolve(env Environment) error {
	err := image.Resolve(env.RunningImages, env.CycleID, env.Monitor, env.Store)
	if err != nil {
		return err
	}
	return nil
}

// Backup calls the Backup function for monitored images
func (f maintainWorkerFunc) Backup(env Environment) error {
	err := image.Backup(env.Store)
	if err != nil {
		return err
	}
	return nil
}

// Restore calls the Restore function for monitored images
func (f maintainWorkerFunc) Restore(env Environment) error {
	err := image.Restore(env.Store)
	if err != nil {
		return err
	}
	return nil
}

// Run worker function as go func
func (mt MaintainTask) Work() {
	err := mt.Func(mt.Env)
	if err != nil {
		log.Error(log.Fields{
			"package":  "main/worker",
			"function": "Run",
			"worker":   "maintainWorker",
			"error":    err.Error(),
		}, "Maintain worker failed")
	}

}
