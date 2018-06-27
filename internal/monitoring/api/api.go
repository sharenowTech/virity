package api

import (
	"net/http"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

type ApiService struct {
	Url     string
	Mux     *http.ServeMux
	Server  *http.Server
	Statics *staticsServer
	Model   Model
}

type Model interface {
	AddImage(image pluginregistry.ImageStack) error
	DelImage(id string) error
	GetImage(id string) ([]byte, error)
	GetImageList() ([]byte, error)
	GetVulnerabilityList() ([]byte, error)
}

func init() {
	// register New function at pluginregistry
	_, err := pluginregistry.RegisterMonitor("internal", New)
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
	api := ApiService{
		Url:     config.Endpoint,
		Mux:     http.NewServeMux(),
		Statics: newStaticsServer("static"),
		Model:   ImageModel{},
	}

	api.Server = &http.Server{
		Addr:    ":8080",
		Handler: api.Mux,
	}

	api.Serve()

	return api
}

func (api ApiService) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
	if status != pluginregistry.StatusOK {
		err := api.Model.AddImage(image)
		if err != nil {
			return err
		}
	}
	return nil
}

func (api ApiService) Resolve(image pluginregistry.ImageStack) error {
	api.Model.DelImage(image.MetaData.ImageID)
	return nil
}

func (api ApiService) Serve() {
	api.router()
	go api.Server.ListenAndServe()
}

func (api ApiService) restartServer() {
	api.Server.Close()
	api.Serve()
}
