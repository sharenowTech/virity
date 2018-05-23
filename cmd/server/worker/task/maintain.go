package task

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
)

type Maintain struct {
	Base    BaseTask
	Process []func(Maintain) error
}

// Resolve resolves monitoring alerts for not running images and resets the list of running images
func Resolve(m Maintain) error {
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
func Backup(m Maintain) error {
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
func Restore(m Maintain) error {
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
func (m Maintain) Run() error {
	for _, proc := range m.Process {
		err := proc(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Maintain) Retry() {
	if m.Base.Retries > 0 {
		return
	}
	m.Base.Retries--
	AddToQueue(m)
	return
}

func (m *Maintain) Register() {
	m.Base.wg.Add(1)
}

func (m *Maintain) DeRegister() {
	m.Base.wg.Done()
}
