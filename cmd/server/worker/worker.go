package worker

import (
	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Environment of the workers with the needed resources etc.
type Environment struct {
	RunningImages *image.Active
	Store         pluginregistry.Store
	Scanner       pluginregistry.Scan
	Monitor       pluginregistry.Monitor
	ErrorRetries  int
	CycleID       int
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

// InitStackWorker initializes a new Stack Worker
func (e Environment) InitCGroupWorker(stack pluginregistry.ContainerGroup) cGroupWorker {
	log.Debug(log.Fields{
		"package":  "main/worker",
		"function": "InitCGroupWorker",
	}, "Stack worker created")
	return cGroupWorker{
		Group: stack,
		env:   e,
	}
}

// InitMaintainWorker initializes a new Maintain Worker
func (e Environment) InitMaintainWorker() maintainWorker {
	log.Debug(log.Fields{
		"package":  "main/worker",
		"function": "InitMaintainWorker",
	}, "Maintain worker created")
	return maintainWorker{
		env: e,
	}
}
