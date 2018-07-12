package api

import (
	"net/http"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
	"github.com/gorilla/mux"
)

// Todo: Alternative for the global state
var defService = APIService{
	Mux:     mux.NewRouter(),
	Statics: newStaticsServer("static"),
	Model:   NewModel(),
}

// APIService holds all necessary server objects
type APIService struct {
	URL     string
	Mux     *mux.Router
	Server  *http.Server
	Statics *staticsServer
	Model   Model
	running bool //is true if the server is already running
}

// Model is an interface of all functionality a model has to provide
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

	defService.URL = config.Endpoint
	defService.Server = &http.Server{
		Addr: defService.URL,
	}

	if defService.running == false {
		defService.Serve()
		defService.running = true
	} else {
		log.Warn(log.Fields{
			"function": "New",
			"package":  "api",
		}, "API Server is already running")
	}

	log.Debug(log.Fields{
		"function": "New",
		"package":  "api",
	}, "API plugin initialized")

	return defService
}

func (api APIService) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
	if status != pluginregistry.StatusOK {
		log.Debug(log.Fields{
			"function": "Push",
			"package":  "api",
		}, "Sending data to internal api")
		err := api.Model.AddImage(image)
		if err != nil {
			return err
		}
	}
	return nil
}

func (api APIService) Resolve(image pluginregistry.ImageStack) error {
	api.Model.DelImage(image.MetaData.ImageID)
	return nil
}

func (api APIService) Serve() {
	api.router()
	go api.Server.ListenAndServe()
}

func (api APIService) restartServer() {
	api.Server.Close()
	api.Serve()
}
