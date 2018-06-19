package api

import (
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

type api struct {
	url string
}

func init() {
	// register New function at pluginregistry
	_, err := pluginregistry.RegisterMonitor("sensu", New)
	if err != nil {
		log.Info(log.Fields{
			"function": "init",
			"package":  "api",
			"error":    err.Error(),
		}, "An error occoured while register a monitor")
	}
}

// New initializes the plugin
func New(config pluginregistry.Config) pluginregistry.Monitor {
	api := api{
		url: config.Endpoint,
	}

	return api
}

func (a api) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
	panic("not implemented")
}

func (a api) Resolve(image pluginregistry.ImageStack) error {
	panic("not implemented")
}
