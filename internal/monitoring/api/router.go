package api

import (
	"github.com/gorilla/handlers"
)

func (api APIService) router() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	api.Server.Handler = handlers.CORS(headersOk, originsOk, methodsOk)(defService.Mux)
	// serve api
	api.Mux.HandleFunc("/api/image/{id}", api.Image)
	api.Mux.HandleFunc("/api/image/", api.ImageList)
	api.Mux.PathPrefix("/").Handler(api.Statics)
}
