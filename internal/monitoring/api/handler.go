package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api APIService) ImageList(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	sendData, _ := api.Model.GetImageList()
	w.Header().Set("Content-Type", "application/json")
	w.Write(sendData)
}

func (api APIService) Image(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	id := mux.Vars(req)
	sendData, _ := api.Model.GetImage(id["id"])
	w.Header().Set("Content-Type", "application/json")
	w.Write(sendData)
}
