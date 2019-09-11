package task

import (
	"github.com/sharenowTech/virity/cmd/server/image"
	"github.com/sharenowTech/virity/internal/log"
)

// maintain Task to maintain the Environment
type maintain struct {
	Base    BaseTask
	Process []func(maintain) error
}

// Resolve resolves monitoring alerts for not running images and resets the list of running images
func Resolve(m maintain) error {
	log.Debug(log.Fields{
		"package":  "worker",
		"function": "Resolve",
	}, "Run task")
	err := image.Resolve(m.Base.RunningImages, m.Base.CycleID, m.Base.Monitor, m.Base.Store)
	if err != nil {
		return err
	}
	return nil
}

// Backup calls the Backup function for monitored images
func Backup(m maintain) error {
	log.Debug(log.Fields{
		"package":  "worker",
		"function": "Backup",
	}, "Run task")
	err := image.Backup(m.Base.Store)
	if err != nil {
		return err
	}
	return nil
}

// Restore calls the Restore function for monitored images
func Restore(m maintain) error {
	log.Debug(log.Fields{
		"package":  "worker",
		"function": "Restore",
	}, "Run task")
	err := image.Restore(m.Base.Store)
	if err != nil {
		return err
	}
	return nil
}

// Run worker function as go func
func (m maintain) Run() error {
	for _, proc := range m.Process {
		err := proc(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// Retry adds the task to the queue until the counter is 0
func (m *maintain) Retry() {
	if m.Base.Retries <= 0 {
		log.Warn(log.Fields{
			"package":  "worker",
			"function": "Maintain/Retry",
		}, "Task failed after several retries. It will be dropped")
		m.Base.wg.Done()
		return
	}
	log.Info(log.Fields{
		"package":  "worker",
		"function": "Maintain/Retry",
	}, "Task failed. I will retry.")
	m.Base.Retries--
	Queue <- m
	return
}

// Register adds Task to a WaitGroup
func (m *maintain) Register() {
	m.Base.wg.Add(1)
}

// DeRegister sets the Task in the WaitGroup to Done
func (m *maintain) DeRegister() {
	m.Base.wg.Done()
}
