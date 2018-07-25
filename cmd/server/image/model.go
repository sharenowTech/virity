package image

import (
	"sync"
)

// Map contains a maps to store monitored and currently active images
type Map struct {
	data *sync.Map
}

// NewMap creates and initializes a new model
func NewMap() *Map {
	return &Map{
		data: &sync.Map{},
	}

}

// Add adds a new image to the model
func (model Map) Add(image Data) {
	add(model.data, image)
}

// Delete removes an image from the model
func (model Map) Delete(image Data) {
	delete(model.data, image)
}

// Read returns the image from the model by a given id (ImageID)
func (model Map) Read(id string) (val Data, ok bool) {
	return read(model.data, id)
}

// Range iterates over the model and applies the given function
func (model Map) Range(f func(key, val interface{}) bool) {
	iterate(model.data, f)
}

// Reset overwrites the model with empty data
func (model Map) Reset() {
	model.data = &sync.Map{}
}

// UpdateState updates the Status of a provided image in the model
func (model Map) UpdateState(state Status, cycleID int, attr Data) Data {
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

func add(list *sync.Map, item Data) {
	list.Store(item.Image.MetaData.ImageID, item)
}

func delete(list *sync.Map, item Data) {
	list.Delete(item.Image.MetaData.ImageID)
}

func read(list *sync.Map, id string) (val Data, ok bool) {
	if img, ok := list.Load(id); ok {
		return img.(Data), true
	}

	return Data{}, false
}

func iterate(list *sync.Map, f func(key, val interface{}) bool) {
	list.Range(f)
}
