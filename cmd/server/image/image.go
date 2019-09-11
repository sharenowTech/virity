package image

import (
	"fmt"
	"path"
	"reflect"
	"sort"
	"sync"

	"github.com/sharenowTech/virity/internal/config"
	"github.com/sharenowTech/virity/internal/log"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

// MonitorStatus is the type for monitoring status based on an integer (to create a enum like structure)
type action int

type status int

const (
	noop action = iota
	update
	partlyResolve
	fullyResolve
)

const (
	running status = iota
	scanning
	scanned
	monitored
	resolved
)

const backupPath = "Backup/Monitored"

// ImageStatus contains the image, the current state of the image (running, scanned or monitored) and if it needs to be resolved or updated
type imageStatus struct {
	image          image
	state          status
	action         action
	stateChangedAt int
}

type image pluginregistry.ImageStack

// Active containers all currently running images (must be stored for each interval separately)
type Active struct {
	images sync.Map
}

// Monitored contains a maps to store monitored and currently active images
type Monitored struct {
	images sync.Map
}

var defMonitored = &Monitored{}

// Scan scans image and returns vulnerabilities
func (i imageStatus) scan(scanner pluginregistry.Scan) (*pluginregistry.Vulnerabilities, error) {
	vuln, scanErr := scanner.Scan(pluginregistry.Image(i.image.MetaData))
	if scanErr != nil {
		return nil, fmt.Errorf("Image: %v - %v", i.image.MetaData.Tag, scanErr.Error())
	}
	return vuln, nil
}

// Monitor pushes the stack to the specified monitor
func (i imageStatus) monitor(monitor pluginregistry.Monitor) error {
	configScan := config.GetScanConfig()
	severity := pluginregistry.VulnSeverity(configScan.SeverityLevel)
	i.image.Vuln.CVE = filterCVEs(severity, i.image.Vuln.CVE)
	log.Info(log.Fields{
		"package":  "main/image",
		"function": "monitor",
		"count":    len(i.image.Vuln.CVE),
		"severity": severity,
		"image":    i.image.MetaData.Tag,
	}, "Vulnerabilities found")

	status := evalStatus(i.image.Vuln.CVE, severity)
	err := monitor.Push(pluginregistry.ImageStack(i.image), status)
	if err != nil {
		return err
	}
	return nil
}

// Resolve pushes the stack to the specified monitor to resolve the issue
func (i imageStatus) resolve(monitor pluginregistry.Monitor) error {
	err := monitor.Resolve(pluginregistry.ImageStack(i.image))
	if err != nil {
		return err
	}
	return nil
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

// Restore all monitored images from store
func Restore(store pluginregistry.Store) error {
	return defMonitored.Restore(store)
}

// Restore all monitored images from store
func (m *Monitored) Restore(store pluginregistry.Store) error {
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

// DataAdd extracts the image from a container and adds it to a map
func dataAdd(data *sync.Map, container pluginregistry.Container) imageStatus {
	if val, ok := data.Load(container.ImageID); ok {
		log.Info(log.Fields{
			"package":  "main/image",
			"function": "dataAdd",
			"image":    container.Image,
			"id":       container.ImageID,
			"owner":    container.OwnerID,
			"state":    val.(imageStatus).state,
		}, "Image exists. I will add the container and owner")
		stack := val.(imageStatus).image
		stack.Containers = appendContainer(stack.Containers, container)
		stack.MetaData.OwnerID = appendOwner(stack.MetaData.OwnerID, container.OwnerID)

		is := imageStatus{
			image: stack,
			state: running,
		}

		data.Store(container.ImageID, is)
		return is
	}
	log.Debug(log.Fields{
		"package":  "main/image",
		"function": "dataAdd",
		"image":    container.Image,
		"id":       container.ImageID,
		"owner":    container.OwnerID,
	}, "Add new image to list")
	containers := make([]pluginregistry.Container, 1)
	containers[0] = container

	owners := appendOwner(make([]string, 0), container.OwnerID)

	stack, _ := data.LoadOrStore(container.ImageID, imageStatus{
		image: image{
			MetaData: pluginregistry.Image{
				ImageID: container.ImageID,
				Tag:     container.Image,
				OwnerID: owners,
			},
			Containers: containers,
		},
		state: running})

	return stack.(imageStatus)
}

// Appends a container to a provided list if it does not already exist
func appendContainer(list []pluginregistry.Container, container pluginregistry.Container) []pluginregistry.Container {
	exist := false

	for _, val := range list {
		if val.ID == container.ID {
			exist = true
		}
	}
	if !exist {
		list = append(list, container)
	}
	return list
}

// Appends a Owner to a provided list if it does not already exist
func appendOwner(list []string, ownerID string) []string {
	exist := false
	for _, val := range list {
		if val == ownerID {
			exist = true
		}
	}
	if !exist {
		list = append(list, ownerID)
	}
	return list
}

func equalContainer(slice1, slice2 []pluginregistry.Container) bool {

	if len(slice1) != len(slice2) {
		return false
	}

	if (slice1 == nil) != (slice2 == nil) {
		return false
	}

	sort.Slice(slice1, func(i, j int) bool { return slice1[i].ID < slice1[j].ID })
	sort.Slice(slice2, func(i, j int) bool { return slice2[i].ID < slice2[j].ID })

	slice2 = slice2[:len(slice1)]
	for index, val := range slice1 {
		if val.ID != slice2[index].ID {
			return false
		}
	}
	return true
}

// Difference returns the elements in slice1 that aren't in slice2
func difference(slice1, slice2 []string) []string {
	mapSlice2 := map[string]bool{}
	for _, elem := range slice2 {
		mapSlice2[elem] = true
	}
	diff := []string{}
	for _, elem := range slice1 {
		if _, ok := mapSlice2[elem]; !ok {
			diff = append(diff, elem)
		}
	}
	return diff
}
