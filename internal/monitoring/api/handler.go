package api

import (
	"net/http"
)

func (api ApiService) ImageList(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	sendData, _ := api.Model.GetImageList()
	w.Write(sendData)
}
