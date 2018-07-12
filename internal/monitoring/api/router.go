package api

import "github.com/gorilla/handlers"

func (api APIService) router() {
	api.Server.Handler = handlers.CORS()(defService.Mux)
	// serve api
	api.Mux.HandleFunc("/api/image/{id}", api.Image)
	api.Mux.HandleFunc("/api/image/", api.ImageList)
	api.Mux.PathPrefix("/").Handler(api.Statics)
}
