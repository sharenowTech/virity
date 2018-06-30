package image

import (
	"fmt"
	"path"
	"reflect"
	"sync"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// Monitored contains a maps to store monitored and currently active images
type Monitored struct {
	images sync.Map
}

var defMonitored = &Monitored{}

// RestoreFrom all monitored images from store
func RestoreFrom(store pluginregistry.Store) error {
	return defMonitored.RestoreFrom(store)
}

// RestoreFrom all monitored images from store
func (m *Monitored) RestoreFrom(store pluginregistry.Store) error {
	stacks, err := store.LoadImageStacks(backupPath)
	if err != nil {
		return err
	}

	for _, img := range stacks {
		m.images.Store(img.MetaData.ImageID, imageStatus{
			image: image(img),
			state: monitored,
		})
	}
	return nil
}

// Backup all monitored images to store
func Backup(store pluginregistry.Store) error {
	return defMonitored.Backup(store)
}

// Backup all monitored images to store
func (m *Monitored) Backup(store pluginregistry.Store) error {
	var err error
	m.images.Range(func(k, v interface{}) bool {
		val := v.(imageStatus)

		if val.state != monitored {
			log.Debug(log.Fields{
				"package":  "main/image",
				"function": "Backup",
				"image":    val.image.MetaData.Tag,
				"state":    val.state,
				"imageID":  val.image.MetaData.ImageID,
			}, "Image is not yet monitored. I will not backup")
			return true
		}

		err = store.StoreImageStack(pluginregistry.ImageStack(val.image), backupPath)
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
func Refresh(monitor pluginregistry.Monitor) error {
	return defMonitored.Refresh(monitor)
}

// Refresh sends all currently monitored images to the monitor.
func (m *Monitored) Refresh(monitor pluginregistry.Monitor) error {
	var err error
	m.images.Range(func(k, v interface{}) bool {
		val := v.(imageStatus)
		err = val.monitor(monitor)
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// Monitor sends the image data to a provided monitor plugin and update its state
func Monitor(stack imageStatus, cycleID int, monitor pluginregistry.Monitor) error {
	return defMonitored.Monitor(stack, cycleID, monitor)
}

// Monitor sends the image data to a provided monitor plugin and update its state
func (m *Monitored) Monitor(stack imageStatus, cycleID int, monitor pluginregistry.Monitor) error {
	err := stack.monitor(monitor)
	if err != nil {
		return err
	}

	err = m.updateState(stack.image.MetaData.ImageID, monitored, cycleID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateMonitoredState updates the state of an image (key is provided) in the monitored map
func (m *Monitored) updateState(key string, state status, cycleID int) error {
	if val, ok := m.images.Load(key); ok {
		img := val.(imageStatus)
		img.state = state
		img.stateChangedAt = cycleID
		m.images.Store(key, img)
		return nil
	}

	return fmt.Errorf("Image %v not found in monitored", key)
}

// Analyse scans container image and returns the image data
func Analyse(container pluginregistry.Container, cycleID int, scanner pluginregistry.Scan) (*imageStatus, error) {
	return defMonitored.Analyse(container, cycleID, scanner)
}

// Analyse scans container image and returns the image data
func (m *Monitored) Analyse(container pluginregistry.Container, cycleID int, scanner pluginregistry.Scan) (*imageStatus, error) {
	log.Debug(log.Fields{
		"package":   "main/image",
		"function":  "Analyse",
		"image":     container.Image,
		"container": container.Name,
		"owner":     container.OwnerID,
		"hostname":  container.Hostname,
	}, "Add image to running list")
	stack := m.add(container)

	if stack.state == scanning {
		log.Debug(log.Fields{
			"package":  "main/image",
			"function": "Analyse",
			"image":    stack.image.MetaData.Tag,
			"id":       stack.image.MetaData.ImageID,
			"owner":    stack.image.MetaData.OwnerID,
		}, "Image is currently being scanned and will therefore not be analysed in this cycle")
		return &stack, nil
	}

	err := m.updateState(stack.image.MetaData.ImageID, scanning, cycleID)
	if err != nil {
		return nil, err
	}

	vuln, err := stack.scan(scanner)
	if err != nil {
		stateErr := m.updateState(stack.image.MetaData.ImageID, running, cycleID)
		if stateErr != nil {
			return nil, err
		}
		return nil, err
	}

	stack.image.Vuln = *vuln
	m.images.Store(stack.image.MetaData.ImageID, stack)

	err = m.updateState(stack.image.MetaData.ImageID, scanned, cycleID)
	if err != nil {
		return nil, err
	}
	return &stack, nil

}

// Resolve compares monitored and active image maps and resolves differences
func Resolve(active *Active, cycleID int, monitor pluginregistry.Monitor, store pluginregistry.Store) error {
	return defMonitored.Resolve(active, cycleID, monitor, store)
}

// Resolve compares monitored and active image maps and resolves differences
func (m *Monitored) Resolve(active *Active, cycleID int, monitor pluginregistry.Monitor, store pluginregistry.Store) error {
	resolvable := m.compare(&active.images, cycleID)
	for _, elem := range resolvable {
		switch elem.action {
		case update:
			err := elem.monitor(monitor)
			if err != nil {
				return err
			}
		case partlyResolve:
			fallthrough
		case fullyResolve:
			err := elem.resolve(monitor)
			if err != nil {
				return err
			}
			m.del(elem)
			store.Delete(path.Join(backupPath, elem.image.MetaData.ImageID))
		default:
			return fmt.Errorf("Invalid state of resolvable image")
		}
	}

	return nil
}

// MonitoredDel removes an image from the monitored map, only if the image is currently monitored
func (m *Monitored) del(is imageStatus) {
	if is.state == monitored {
		log.Info(log.Fields{
			"package":  "main/image",
			"function": "del",
			"image":    is.image.MetaData.Tag,
			"id":       is.image.MetaData.ImageID,
			"owner":    is.image.MetaData.OwnerID,
		}, "Deleting Image from monitored list")
		m.images.Delete(is.image.MetaData.ImageID)
		return
	}
	log.Info(log.Fields{
		"package":  "main/image",
		"function": "del",
		"image":    is.image.MetaData.Tag,
		"id":       is.image.MetaData.ImageID,
		"owner":    is.image.MetaData.OwnerID,
		"state":    is.state,
	}, "Image is not yet monitored. Therefore it cannot be deleted.")
}

// MonitoredAdd adds an image to the monitored map based on a provided container
func (m *Monitored) add(container pluginregistry.Container) imageStatus {
	return dataAdd(&m.images, container)
}

// Compare compares the monitored and active list and returns all images which should be resolved/updated --> only Images with state "monitored" are considered
func (m *Monitored) compare(active *sync.Map, cycleID int) []imageStatus {
	different := make([]imageStatus, 0)
	m.images.Range(func(k, v interface{}) bool {
		// If image is not monitored skip
		if v.(imageStatus).state != monitored {
			log.Debug(log.Fields{
				"package":  "main/image",
				"function": "compare",
				"image":    v.(imageStatus).image.MetaData.Tag,
				"id":       v.(imageStatus).image.MetaData.ImageID,
				"owner":    v.(imageStatus).image.MetaData.OwnerID,
			}, "I will not check if Image is resolvable, as it is not yet monitored")
			return true
		}

		if v.(imageStatus).stateChangedAt > cycleID {
			log.Debug(log.Fields{
				"package":       "main/image",
				"function":      "compare",
				"image":         v.(imageStatus).image.MetaData.Tag,
				"id":            v.(imageStatus).image.MetaData.ImageID,
				"owner":         v.(imageStatus).image.MetaData.OwnerID,
				"cycle_changed": v.(imageStatus).stateChangedAt,
				"current_cycle": cycleID,
			}, "Image will not be resolved because the current cycleID is too old")
			return true
		}

		if val, ok := active.Load(k); ok {
			mon := v.(imageStatus).image
			act := val.(imageStatus).image

			// Partial Resolve if some Owner do not exist anymore
			if eq := reflect.DeepEqual(mon.MetaData.OwnerID, act.MetaData.OwnerID); !eq {
				missingOwner := difference(mon.MetaData.OwnerID, act.MetaData.OwnerID)
				mon.MetaData.OwnerID = missingOwner

				value := v.(imageStatus)
				value.image = mon
				value.action = partlyResolve
				different = append(different, value)

				log.Info(log.Fields{
					"package":  "main/image",
					"function": "compare",
					"image":    mon.MetaData.Tag,
					"id":       mon.MetaData.ImageID,
					"owner":    mon.MetaData.OwnerID,
					"state":    v.(imageStatus).state,
					"action":   partlyResolve,
				}, "Partly resolve Image")
				return true
			}

			// Only update Monitoring Data if the containers have changed
			if eq := equalContainer(mon.Containers, act.Containers); eq {
				return true
			}

			mon.Containers = act.Containers
			value := v.(imageStatus)
			value.image = mon
			value.action = update
			different = append(different, value)

			log.Info(log.Fields{
				"package":  "main/image",
				"function": "compare",
				"image":    mon.MetaData.Tag,
				"id":       mon.MetaData.ImageID,
				"owner":    mon.MetaData.OwnerID,
				"state":    v.(imageStatus).state,
				"action":   update,
			}, "Update Image")
			return true

		}

		// fully resolve image
		value := v.(imageStatus)
		value.action = fullyResolve
		different = append(different, value)
		log.Info(log.Fields{
			"package":  "main/image",
			"function": "compare",
			"image":    value.image.MetaData.Tag,
			"id":       value.image.MetaData.ImageID,
			"owner":    value.image.MetaData.OwnerID,
			"state":    v.(imageStatus).state,
			"action":   fullyResolve,
		}, "Fully resolve Image")
		return true
	})

	return different
}
