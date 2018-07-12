package api

import (
	"encoding/json"
	"fmt"
	"sync"

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

type cacheModel struct {
	ID       string   `json:"id"`
	Name     string   `json:"tag"`
	Owner    []string `json:"owner"`
	CveCount int      `json:"cve_count"`
}

// ImageModel holds the API data
type ImageModel struct {
	images  map[string]image
	updated bool
	cache   []cacheModel
	mutex   *sync.Mutex
}

// NewModel initializes the ImageModel
// returns a pointer to the model
func NewModel() *ImageModel {
	return &ImageModel{
		images: make(map[string]image),
		mutex:  &sync.Mutex{},
	}
}

// AddImage adds a new image to the API (monitor)
func (im ImageModel) AddImage(stack pluginregistry.ImageStack) error {
	err := im.update(func(stack interface{}) error {
		if val, ok := stack.(pluginregistry.ImageStack); ok {
			im.images[val.MetaData.ImageID] = image{
				ID:         val.MetaData.ImageID,
				Tag:        val.MetaData.Tag,
				Containers: val.Containers,
				CVEs:       val.Vuln.CVE,
				Owner:      val.MetaData.OwnerID,
			}
			return nil
		}
		return fmt.Errorf("Type Error: %v is not an ImageStack", stack)
	}, stack)
	if err != nil {
		return err
	}
	return nil
}

// DelImage removes an image from the API (monitor)
func (im ImageModel) DelImage(id string) error {
	err := im.update(func(id interface{}) error {
		if val, ok := id.(string); ok {
			delete(im.images, val)
			return nil
		}
		return fmt.Errorf("Type Error: %v is not a string", id)
	}, id)
	if err != nil {
		return err
	}
	return nil
}

// GetImage returns an image to the API (monitor)
func (im ImageModel) GetImage(id string) ([]byte, error) {
	if _, ok := im.images[id]; !ok {
		return nil, fmt.Errorf("Image not found")
	}
	return toJSON(im.images[id])
}

// GetImageList returns all images to the API (monitor)
func (im ImageModel) GetImageList() ([]byte, error) {
	if im.updated == true || im.cache == nil {
		imageList := make([]cacheModel, len(im.images))
		counter := 0
		for _, img := range im.images {
			imageList[counter] = cacheModel{
				ID:       img.ID,
				Name:     img.Tag,
				Owner:    img.Owner,
				CveCount: len(img.CVEs),
			}
			counter++
		}
		err := im.update(func(list interface{}) error {
			if val, ok := list.([]cacheModel); ok {
				im.cache = val
				return nil
			}
			return fmt.Errorf("Type Error: %v is not a []cacheModel", list)
		}, imageList)
		if err != nil {
			return nil, err
		}
	}
	return toJSON(im.cache)
}

// GetVulnerabilityList returns all vulnerabilities to the API (monitor)
func (im ImageModel) GetVulnerabilityList() ([]byte, error) {
	panic("not implemented")
}

func (im ImageModel) update(action func(interface{}) error, param interface{}) error {
	im.mutex.Lock()
	err := action(param)
	if err != nil {
		im.mutex.Unlock()
		return err
	}
	im.updated = true
	im.mutex.Unlock()
	return nil
}

func toJSON(obj interface{}) ([]byte, error) {

	json, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}
	return json, nil
}
