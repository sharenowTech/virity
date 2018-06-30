package image

import (
	"fmt"
	"sync"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Active containers all currently running images (must be stored for each interval separately)
type Active struct {
	images sync.Map
}

// Running adds the image of the container to the currently running list
/*func Running(container pluginregistry.Container) error {
	return defData.Running(container)
}*/

func (a *Active) Get(key string) (imageStatus, error) {
	if val, ok := a.images.Load(key); ok {
		return val.(imageStatus), nil
	}
	return imageStatus{}, fmt.Errorf("Image not found")
}

// Add adds the image of the container to the currently running list
func (a *Active) Add(container pluginregistry.Container) error {
	dataAdd(&a.images, container)
	return nil
}

// ResetActive resets the map of active containers
/*func ResetActive() {
	defMonitored.ResetActive()
}*/

// Reset resets the map of active containers
func (a *Active) Reset() {
	a.images = sync.Map{}
}

// ActiveDel removes an image from the active map independed of its state as it is not necessary
func (a *Active) del(is imageStatus) {
	log.Info(log.Fields{
		"package":  "main/image",
		"function": "del",
		"image":    is.image.MetaData.Tag,
		"id":       is.image.MetaData.ImageID,
		"owner":    is.image.MetaData.OwnerID,
		"state":    is.state,
	}, "Deleting Image from active list")
	a.images.Delete(is.image.MetaData.ImageID)
}
