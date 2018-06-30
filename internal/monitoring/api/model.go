package api

import (
	"encoding/json"
	"fmt"

	"github.com/car2go/virity/internal/pluginregistry"
)

type image struct {
	ArrayID    int
	ID         string                     `json:"id"`
	Tag        string                     `json:"tag"`
	Containers []pluginregistry.Container `json:"in_containers"`
	CVEs       []pluginregistry.CVE       `json:"vulnerability_cve"`
	Owner      []string                   `json:"owner"`
}

type meta struct {
	ID       string   `json:"id"`
	Name     string   `json:"tag"`
	Owner    []string `json:"owner"`
	CveCount int      `json:"cve_count"`
}

type ImageModel struct {
	images map[string]image
	meta   []meta
}

func NewModel() *ImageModel {
	return &ImageModel{
		images: make(map[string]image),
		meta:   make([]meta, 0, 50),
	}
}

func (im *ImageModel) AddImage(stack pluginregistry.ImageStack) error {
	if _, ok := im.images[stack.MetaData.ImageID]; ok {
		return nil
	}
	im.images[stack.MetaData.ImageID] = image{
		ArrayID:    len(im.meta) + 1,
		ID:         stack.MetaData.ImageID,
		Tag:        stack.MetaData.Tag,
		Containers: stack.Containers,
		CVEs:       stack.Vuln.CVE,
		Owner:      stack.MetaData.OwnerID,
	}
	im.meta = append(im.meta, meta{
		ID:       stack.MetaData.ImageID,
		Name:     stack.MetaData.Tag,
		Owner:    stack.MetaData.OwnerID,
		CveCount: len(stack.Vuln.CVE),
	})
	return nil
}

func (im *ImageModel) DelImage(id string) error {
	if img, ok := im.images[id]; ok {
		im.meta = append(im.meta[:img.ArrayID], im.meta[img.ArrayID+1:]...)
	}
	delete(im.images, id)
	im.reinit()
	return nil
}

func (im *ImageModel) GetImage(id string) ([]byte, error) {
	if img, ok := im.images[id]; ok {
		return toJSON(img)
	}

	return nil, fmt.Errorf("Image not found")
}

func (im *ImageModel) GetImageList() ([]byte, error) {
	return toJSON(im.meta)
}

func (im *ImageModel) GetVulnerabilityList() ([]byte, error) {
	panic("not implemented")
}

func toJSON(obj interface{}) ([]byte, error) {

	json, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}
	return json, nil
}

func (im *ImageModel) reinit() {
	for id, image := range im.meta {
		tmp := im.images[image.ID]
		tmp.ArrayID = id
		im.images[image.ID] = tmp
	}
}
