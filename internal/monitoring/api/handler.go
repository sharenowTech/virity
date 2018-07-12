package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api APIService) ImageList(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	sendData, err := api.Model.GetImageList()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte("{\"Error\": \"An error occoured. Please try again later.\"}"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(sendData)
}

func (api APIService) Image(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	id := mux.Vars(req)
	sendData, err := api.Model.GetImage(id["id"])
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte("{\"Error\": \"An error occoured. Please try again later.\"}"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(sendData)
}
