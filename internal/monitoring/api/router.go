package api

func (api APIService) router() {

	// serve api
	api.Mux.HandleFunc("/api/image/{id}", api.Image)
	api.Mux.HandleFunc("/api/image/", api.ImageList)
	api.Mux.PathPrefix("/").Handler(api.Statics)
}
