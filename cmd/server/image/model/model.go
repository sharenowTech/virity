package model

import (
	"sync"

	"github.com/car2go/virity/cmd/server/image"
)

// ImageMap contains a maps to store monitored and currently active images
type ImageMap struct {
	data *sync.Map
}

// New creates and initializes a new model
func New() *ImageMap {
	return &ImageMap{
		data: &sync.Map{},
	}

}

// Add adds a new image to the model
func (model ImageMap) Add(image image.Data) {
	add(model.data, image)
}

// Delete removes an image from the model
func (model ImageMap) Delete(image image.Data) {
	delete(model.data, image)
}

// Read returns the image from the model by a given id (ImageID)
func (model ImageMap) Read(id string) (val image.Data, ok bool) {
	return read(model.data, id)
}

// Range iterates over the model and applies the given function
func (model ImageMap) Range(f func(key, val interface{}) bool) {
	iterate(model.data, f)
}

// Reset overwrites the model with empty data
func (model ImageMap) Reset() {
	model.data = &sync.Map{}
}

// UpdateState updates the Status of a provided image in the model
func (model ImageMap) UpdateState(state image.Status, cycleID int, attr image.Data) image.Data {
	// If this image exists in the list, update
	if image, ok := read(model.data, attr.Image.MetaData.ImageID); ok {
		image.State = state
		image.StateChangedAt = cycleID
		add(model.data, image)
		return image
	}

	attr.State = state
	attr.StateChangedAt = cycleID
	return attr
}

func add(list *sync.Map, item image.Data) {
	list.Store(item.Image.MetaData.ImageID, item)
}

func delete(list *sync.Map, item image.Data) {
	list.Delete(item.Image.MetaData.ImageID)
}

func read(list *sync.Map, id string) (val image.Data, ok bool) {
	if img, ok := list.Load(id); ok {
		return img.(image.Data), true
	}

	return image.Data{}, false
}

func iterate(list *sync.Map, f func(key, val interface{}) bool) {
	list.Range(f)
}
