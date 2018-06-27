package api

func (api ApiService) router() {

	// serve api
	api.Mux.Handle("/", api.Statics)
	api.Mux.HandleFunc("/api/", api.handler1)
}
