package api

import (
	"net/http"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Todo: Alternative for the global state
var runningServices = make(map[string]Service)

// Service holds all necessary server objects
type Service struct {
	Mux     *mux.Router
	Server  *http.Server
	Statics *staticsServer
	Model   Model
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

	if service, ok := runningServices[config.Endpoint]; ok {
		log.Info(log.Fields{
			"function": "New",
			"package":  "api",
			"endpoint": config.Endpoint,
		}, "API service is already running on specified endpoint")
		return service
	}

	service := Service{
		Mux:     mux.NewRouter(),
		Statics: newStaticsServer("static"),
		Model:   NewModel(),
		Server: &http.Server{
			Addr: config.Endpoint,
		},
	}

	service.Serve()
	runningServices[config.Endpoint] = service

	log.Debug(log.Fields{
		"function": "New",
		"package":  "api",
	}, "API plugin initialized")

	return service
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

	api.Server.Handler = handlers.CORS(headersOk, originsOk, methodsOk)(api.Mux)
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

func (api Service) StopServer() {
	endpoint := api.Server.Addr
	if _, ok := runningServices[endpoint]; !ok {
		return
	}

	delete(runningServices, endpoint)
	api.Server.Close()

}
