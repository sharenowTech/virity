package image

import (
	"fmt"
	"path"
	"reflect"

	"github.com/car2go/virity/cmd/server/image/model"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

type Model interface {
	Add(image model.ImageStatus)
	Delete(image model.ImageStatus)
	Read(id string) (val model.ImageStatus, ok bool)
	Range() func(f func(key, val interface{}) bool)
	Reset()
	UpdateState(state model.Status, cycleID int, attr model.ImageStatus) model.ImageStatus
}

// RestoreFrom all monitored images from store
func RestoreFrom(store pluginregistry.Store, m Model) error {
	stacks, err := store.LoadImageStacks(backupPath)
	if err != nil {
		return err
	}

	for _, img := range stacks {
		m.Add(model.ImageStatus{
			Image: model.Image(img),
			State: model.Monitored,
		})
	}
	return nil
}

// Backup all monitored images to store
func Backup(store pluginregistry.Store, m Model) error {
	var err error
	m.Range()(func(k, v interface{}) bool {
		val := v.(model.ImageStatus)

		if val.State != model.Monitored {
			log.Debug(log.Fields{
				"package":  "main/image",
				"function": "Backup",
				"image":    val.Image.MetaData.Tag,
				"state":    val.State,
				"imageID":  val.Image.MetaData.ImageID,
			}, "Image is not yet monitored. I will not backup")
			return true
		}

		err = store.StoreImageStack(pluginregistry.ImageStack(val.Image), backupPath)
		if err != nil {
			return false
		}

		return true
	})
	if err != nil {
		return err
	}

	return nil
}

// Refresh sends all currently monitored images to the monitor.
func Refresh(monitor pluginregistry.Monitor, m Model) error {
	var err error
	m.Range()(func(k, v interface{}) bool {
		val := v.(model.ImageStatus)
		err = val.Monitor(monitor)
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// Monitor sends the image data to a provided monitor plugin and update its state
func Monitor(image model.ImageStatus, cycleID int, monitor pluginregistry.Monitor, m Model) error {
	err := image.Monitor(monitor)
	if err != nil {
		return err
	}

	m.UpdateState(model.Monitored, cycleID, image)

	return nil
}

// Analyse scans container image
// The data is persisted in the monitored model
func Analyse(container pluginregistry.Container, cycleID int, scanner pluginregistry.Scan, m Model) (val model.ImageStatus, analysed bool, err error) {
	log.Debug(log.Fields{
		"package":   "main/image",
		"function":  "Analyse",
		"image":     container.Image,
		"container": container.Name,
		"owner":     container.OwnerID,
		"hostname":  container.Hostname,
	}, "Persist current image")

	var image model.ImageStatus
	if exists, ok := m.Read(container.ImageID); ok {
		image = exists
	} else {
		image = model.ImageStatus{}
	}
	image = model.CreateImageStatus(container, image)

	if image.State == model.Scanning {
		log.Debug(log.Fields{
			"package":  "main/image",
			"function": "Analyse",
			"image":    image.Image.MetaData.Tag,
			"id":       image.Image.MetaData.ImageID,
			"owner":    image.Image.MetaData.OwnerID,
		}, "Image is currently being scanned and will therefore not be analysed in this cycle")
		return image, false, nil
	}

	image = m.UpdateState(model.Scanning, cycleID, image)
	vuln, err := image.Scan(scanner)
	if err != nil {
		m.UpdateState(model.Running, cycleID, image)
		return image, false, err
	}
	image.Image.Vuln = *vuln

	m.Add(image)
	image = m.UpdateState(model.Scanned, cycleID, image)

	return image, true, nil
}

// Resolve compares monitored and active image maps and resolves differences
func Resolve(monitored, active Model, cycleID int, monitor pluginregistry.Monitor, store pluginregistry.Store) error {
	resolvable := compare(monitored, active, cycleID)
	for _, elem := range resolvable {
		switch elem.Action {
		case model.Update:
			err := elem.Monitor(monitor)
			if err != nil {
				return err
			}
		case model.PartlyResolve:
			fallthrough
		case model.FullyResolve:
			err := elem.Resolve(monitor)
			if err != nil {
				return err
			}
			monitored.Delete(elem)
			store.Delete(path.Join(backupPath, elem.Image.MetaData.ImageID))
		default:
			return fmt.Errorf("Invalid state of resolvable image")
		}
	}

	return nil
}

// Delete removes an image from the monitored map, only if the image is currently monitored
func Delete(is model.ImageStatus, force bool, m Model) {
	if is.State == model.Monitored || force {
		log.Info(log.Fields{
			"package":  "main/image",
			"function": "del",
			"image":    is.Image.MetaData.Tag,
			"id":       is.Image.MetaData.ImageID,
			"owner":    is.Image.MetaData.OwnerID,
		}, "Deleting Image from model list")
		m.Delete(is)
		return
	}
	log.Info(log.Fields{
		"package":  "main/image",
		"function": "del",
		"image":    is.Image.MetaData.Tag,
		"id":       is.Image.MetaData.ImageID,
		"owner":    is.Image.MetaData.OwnerID,
		"state":    is.State,
	}, "Image is not yet monitored. Therefore it cannot be deleted.")
}

// Add adds an image to the monitored map based on a provided container
func Add(container pluginregistry.Container, m Model) model.ImageStatus {
	var image model.ImageStatus
	if val, ok := m.Read(container.ImageID); ok {
		log.Debug(log.Fields{
			"package":  "main/image",
			"function": "Add",
			"image":    container.Image,
			"id":       container.ImageID,
			"owner":    container.OwnerID,
			"state":    val.State,
		}, "Image exists. I will update the data")

		image = model.CreateImageStatus(container, val)
	}

	image = model.CreateImageStatus(container, model.ImageStatus{})
	m.Add(image)
	return image
}

// Reset overwrites the current Model with a new one
func Reset(m Model) {
	m.Reset()
}

// Read returns a value based on a key
func Read(key string, m Model) (val model.ImageStatus, ok bool) {
	return m.Read(key)
}

// Compare compares the monitored and active list and returns all images which should be resolved/updated --> only Images with state "monitored" are considered
func compare(monitored, active Model, cycleID int) []model.ImageStatus {
	different := make([]model.ImageStatus, 0)
	monitored.Range()(func(k, v interface{}) bool {
		// If image is not monitored skip
		if v.(model.ImageStatus).State != model.Monitored {
			log.Debug(log.Fields{
				"package":  "main/image",
				"function": "compare",
				"image":    v.(model.ImageStatus).Image.MetaData.Tag,
				"id":       v.(model.ImageStatus).Image.MetaData.ImageID,
				"owner":    v.(model.ImageStatus).Image.MetaData.OwnerID,
			}, "I will not check if Image is resolvable, as it is not yet monitored")
			return true
		}

		if v.(model.ImageStatus).StateChangedAt > cycleID {
			log.Debug(log.Fields{
				"package":       "main/image",
				"function":      "compare",
				"image":         v.(model.ImageStatus).Image.MetaData.Tag,
				"id":            v.(model.ImageStatus).Image.MetaData.ImageID,
				"owner":         v.(model.ImageStatus).Image.MetaData.OwnerID,
				"cycle_changed": v.(model.ImageStatus).StateChangedAt,
				"current_cycle": cycleID,
			}, "Image will not be resolved because the current cycleID is too old")
			return true
		}

		if val, ok := active.Read(k.(string)); ok {
			mon := v.(model.ImageStatus).Image
			act := val.Image

			// Partial Resolve if some Owner do not exist anymore
			if eq := reflect.DeepEqual(mon.MetaData.OwnerID, act.MetaData.OwnerID); !eq {
				missingOwner := difference(mon.MetaData.OwnerID, act.MetaData.OwnerID)
				mon.MetaData.OwnerID = missingOwner

				value := v.(model.ImageStatus)
				value.Image = mon
				value.Action = model.PartlyResolve
				different = append(different, value)

				log.Info(log.Fields{
					"package":  "main/image",
					"function": "compare",
					"image":    mon.MetaData.Tag,
					"id":       mon.MetaData.ImageID,
					"owner":    mon.MetaData.OwnerID,
					"state":    v.(model.ImageStatus).State,
					"action":   model.PartlyResolve,
				}, "Partly resolve Image")
				return true
			}

			// Only update Monitoring Data if the containers have changed
			if eq := equalContainer(mon.Containers, act.Containers); eq {
				return true
			}

			mon.Containers = act.Containers
			value := v.(model.ImageStatus)
			value.Image = mon
			value.Action = model.Update
			different = append(different, value)

			log.Info(log.Fields{
				"package":  "main/image",
				"function": "compare",
				"image":    mon.MetaData.Tag,
				"id":       mon.MetaData.ImageID,
				"owner":    mon.MetaData.OwnerID,
				"state":    v.(model.ImageStatus).State,
				"action":   model.Update,
			}, "Update Image")
			return true

		}

		// fully resolve image
		value := v.(model.ImageStatus)
		value.Action = model.FullyResolve
		different = append(different, value)
		log.Info(log.Fields{
			"package":  "main/image",
			"function": "compare",
			"image":    value.Image.MetaData.Tag,
			"id":       value.Image.MetaData.ImageID,
			"owner":    value.Image.MetaData.OwnerID,
			"state":    v.(model.ImageStatus).State,
			"action":   model.FullyResolve,
		}, "Fully resolve Image")
		return true
	})

	return different
}
