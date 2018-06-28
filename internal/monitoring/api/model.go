package api

import (
	"encoding/json"
	"fmt"

	"github.com/car2go/virity/internal/pluginregistry"
)

type image struct {
	ID         string                     `json:"id"`
	Tag        string                     `json:"tag"`
	Containers []pluginregistry.Container `json:"in_containers"`
	CVEs       []pluginregistry.CVE       `json:"vulnerability_cve"`
	Owner      []string                   `json:"owner"`
}

type ImageModel struct {
	index  map[string]int
	images []image
}

func NewModel() *ImageModel {
	return &ImageModel{
		index:  make(map[string]int),
		images: make([]image, 0, 50),
	}
}

func (im *ImageModel) AddImage(stack pluginregistry.ImageStack) error {
	if _, ok := im.index[stack.MetaData.ImageID]; ok {
		return nil
	}
	im.index[stack.MetaData.ImageID] = len(im.images) + 1
	im.images = append(im.images, image{
		ID:         stack.MetaData.ImageID,
		Tag:        stack.MetaData.Tag,
		Containers: stack.Containers,
		CVEs:       stack.Vuln.CVE,
		Owner:      stack.MetaData.OwnerID,
	})
	return nil
}

func (im *ImageModel) DelImage(id string) error {
	if imgID, ok := im.index[id]; ok {
		im.images = append(im.images[:imgID], im.images[imgID+1:]...)
	}
	delete(im.index, id)
	im.reinit()
	return nil
}

func (im *ImageModel) GetImage(id string) ([]byte, error) {
	if imgID, ok := im.index[id]; ok {
		return toJSON(im.images[imgID])
	}

	return nil, fmt.Errorf("Image not found")
}

func (im *ImageModel) GetImageList() ([]byte, error) {
	return toJSON(im.images)
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
	for id, image := range im.images {
		im.index[image.ID] = id
	}
}
