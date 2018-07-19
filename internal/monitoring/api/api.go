package api

import (
	"net/http"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Todo: Alternative for the global state
var defService = Service{
	Mux:     mux.NewRouter(),
	Statics: newStaticsServer("static"),
	Model:   NewModel(),
}

// Service holds all necessary server objects
type Service struct {
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

// Push adds a new image to model
func (api Service) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
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

// Resolve deletes an image from the model
func (api Service) Resolve(image pluginregistry.ImageStack) error {
	api.Model.DelImage(image.MetaData.ImageID)
	return nil
}

// Serve sets routes and starts server
func (api Service) Serve() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	api.Server.Handler = handlers.CORS(headersOk, originsOk, methodsOk)(defService.Mux)
	//api.Server.Handler = defService.Mux
	// serve api
	api.Mux.HandleFunc("/api/image/{id}", api.Image)
	api.Mux.HandleFunc("/api/image/", api.ImageList)
	api.Mux.PathPrefix("/").Handler(api.Statics)

	go func() {
		err := api.Server.ListenAndServe()
		if err != nil {
			log.Error(log.Fields{
				"function": "Serve",
				"package":  "api",
				"error":    err,
			}, "Failed to serve API Server")
		}
	}()
}

func (api Service) restartServer() {
	api.Server.Close()
	api.Serve()
}
