package api

import (
	"net/http"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

const static = "static"

type api struct {
	url    string
	mux    *http.ServeMux
	server *http.Server
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
	api := api{
		url: config.Endpoint,
		mux: http.NewServeMux(),
	}

	api.server = &http.Server{
		Addr:    ":8080",
		Handler: api.mux,
	}

	api.runServer()

	return api
}

func (a api) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
	panic("not implemented")
}

func (a api) Resolve(image pluginregistry.ImageStack) error {
	panic("not implemented")
}

func (api api) initRoutes() {
	// serve static files
	fs := http.FileServer(http.Dir(static))
	api.mux.Handle("/", http.StripPrefix("/", fs))

	// serve api
	api.mux.HandleFunc("/api/", handler1)
}

func (api api) runServer() {
	api.initRoutes()
	go api.server.ListenAndServe()
}

func (api api) restartServer() {
	api.server.Close()
	api.runServer()
}
