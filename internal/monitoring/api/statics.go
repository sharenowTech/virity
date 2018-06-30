package api

import "net/http"

type staticsServer struct {
	assetsDir string
}

func newStaticsServer(assetsDir string) *staticsServer {

	srv := staticsServer{
		assetsDir: assetsDir,
	}

	return &srv
}

func (s staticsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	handler := http.FileServer(http.Dir(s.assetsDir))

	handler.ServeHTTP(w, r)
}
