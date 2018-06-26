package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/car2go/virity/internal/pluginregistry"
)

func TestNew(T *testing.T) {
	New(pluginregistry.Config{})
	time.Sleep(2 * time.Hour)
	request, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		T.Error(err)
		return
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		T.Error(err)
		return
	}

	if response.StatusCode != 200 {
		T.Errorf("Server not reachable. Code: %v", response.StatusCode)
	}

}
func TestPush(t *testing.T) {
}

func TestResolve(t *testing.T) {
}
