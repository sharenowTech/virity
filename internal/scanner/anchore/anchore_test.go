package anchore

import (
	"testing"

	"github.com/sharenowTech/virity/internal/config"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

func TestScan(t *testing.T) {
	cfg := config.GetScanConfig()
	config := pluginregistry.Config{
		User:     cfg.Username,
		Password: cfg.Password,
		Endpoint: cfg.Endpoint,
	}
	anchore := New(config)

	data := pluginregistry.Image{
		Tag: "kaitsh/ubuntu:latest",
	}

	resp, err := anchore.Scan(data)
	if err != nil {
		t.Error(err)
		return
	}

	if resp.CVE == nil {
		t.Errorf("Somehow there are no cves")
	}

	if resp.Digest != "sha256:e98ebbed9891ba1633d91818fdbb1f5487ccb1e5af00b2284fbace99d16df911" {
		t.Errorf("Wrong Image Digest: %s", resp.Digest)
	}
}

func TestPushImage(t *testing.T) {
	cfg := config.GetScanConfig()
	api := api{
		username: cfg.Username,
		password: cfg.Password,
		endpoint: cfg.Endpoint,
	}
	image := &image{
		Fulltag: "ubuntu:latest",
	}

	image, err := api.PushImage(*image)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestHealthcheck(t *testing.T) {
	cfg := config.GetScanConfig()
	api := api{
		username: cfg.Username,
		password: cfg.Password,
		endpoint: cfg.Endpoint,
	}
	err := api.Healthcheck()
	if err != nil {
		t.Error(err)
	}
}
