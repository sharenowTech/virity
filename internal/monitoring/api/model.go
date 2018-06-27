package api

import (
	"encoding/json"

	"github.com/car2go/virity/internal/pluginregistry"
)

type image struct {
	Tag        string                     `json:"tag"`
	Containers []pluginregistry.Container `json:"in_containers"`
	CVEs       []pluginregistry.CVE       `json:"vulnerability_cve"`
	Owner      []string                   `json:"owner"`
}

type ImageModel struct {
	Data map[string]image
}

func (im ImageModel) AddImage(stack pluginregistry.ImageStack) error {
	im.Data[stack.MetaData.ImageID] = image{
		Tag:        stack.MetaData.Tag,
		Containers: stack.Containers,
		CVEs:       stack.Vuln.CVE,
		Owner:      stack.MetaData.OwnerID,
	}
	return nil
}

func (im ImageModel) DelImage(id string) error {
	delete(im.Data, id)
	return nil
}

func (im ImageModel) GetImage(id string) ([]byte, error) {

	image := im.Data[id]

	return toJSON(image)
}

func (im ImageModel) GetImageList() ([]byte, error) {
	return toJSON(im.Data)
}

func (im ImageModel) GetVulnerabilityList() ([]byte, error) {
	panic("not implemented")
}

func toJSON(obj interface{}) ([]byte, error) {

	json, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}
	return json, nil
}
