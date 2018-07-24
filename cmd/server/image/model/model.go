package model

import (
	"fmt"
	"sync"

	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

// MonitorStatus is the type for monitoring status based on an integer (to create a enum like structure)
type Action int

type Status int

const (
	noop Action = iota
	Update
	PartlyResolve
	FullyResolve
)

const (
	Running Status = iota
	Scanning
	Scanned
	Monitored
	Resolved
)

// ImageStatus contains the image, the current state of the image (running, scanned or monitored) and if it needs to be resolved or updated
type ImageStatus struct {
	Image          Image
	State          Status
	Action         Action
	StateChangedAt int
}

type Image pluginregistry.ImageStack

// Monitored contains a maps to store monitored and currently active images
type Model struct {
	images *sync.Map
}

func New() *Model {
	return &Model{
		images: &sync.Map{},
	}

}

func (model Model) Add(image ImageStatus) {
	add(model.images, image)
}

func (model Model) Delete(image ImageStatus) {
	delete(model.images, image)
}

func (model Model) Read(id string) (val ImageStatus, ok bool) {
	return read(model.images, id)
}

func (model Model) Range() func(f func(key, val interface{}) bool) {
	return iterate(model.images)
}

func (model Model) Reset() {
	model.images = &sync.Map{}
}

func (model Model) UpdateState(state Status, cycleID int, attr ImageStatus) ImageStatus {
	// If this image exists in the list, update
	if image, ok := read(model.images, attr.Image.MetaData.ImageID); ok {
		image.State = state
		image.StateChangedAt = cycleID
		add(model.images, image)
		return image
	}

	attr.State = state
	attr.StateChangedAt = cycleID
	return attr
}

func add(list *sync.Map, item ImageStatus) {
	list.Store(item.Image.MetaData.ImageID, item)
}

func delete(list *sync.Map, item ImageStatus) {
	list.Delete(item.Image.MetaData.ImageID)
}

func read(list *sync.Map, id string) (val ImageStatus, ok bool) {
	if image, ok := list.Load(id); ok {
		return image.(ImageStatus), true
	}

	return ImageStatus{}, false
}

func iterate(list *sync.Map) func(f func(key, val interface{}) bool) {
	return func(f func(key, val interface{}) bool) {
		list.Range(f)
	}
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

// Scan scans image and returns vulnerabilities
func (i ImageStatus) Scan(scanner pluginregistry.Scan) (*pluginregistry.Vulnerabilities, error) {
	vuln, scanErr := scanner.Scan(pluginregistry.Image(i.Image.MetaData))
	if scanErr != nil {
		return nil, fmt.Errorf("Image: %v - %v", i.Image.MetaData.Tag, scanErr.Error())
	}
	return vuln, nil
}

// Monitor pushes the stack to the specified monitor
func (i ImageStatus) Monitor(monitor pluginregistry.Monitor) error {
	configScan := config.GetScanConfig()
	severity := pluginregistry.VulnSeverity(configScan.SeverityLevel)
	i.Image.Vuln.CVE = filterCVEs(severity, i.Image.Vuln.CVE)
	log.Info(log.Fields{
		"package":  "main/image",
		"function": "monitor",
		"count":    len(i.Image.Vuln.CVE),
		"severity": severity,
		"image":    i.Image.MetaData.Tag,
	}, "Vulnerabilities found")

	status := evalStatus(i.Image.Vuln.CVE, severity)
	err := monitor.Push(pluginregistry.ImageStack(i.Image), status)
	if err != nil {
		return err
	}
	return nil
}

// Resolve pushes the stack to the specified monitor to resolve the issue
func (i ImageStatus) Resolve(monitor pluginregistry.Monitor) error {
	err := monitor.Resolve(pluginregistry.ImageStack(i.Image))
	if err != nil {
		return err
	}
	return nil
}

func (i ImageStatus) GetStatus() Status {
	return i.State
}

func (i ImageStatus) GetData() ImageStatus {
	return i
}

// CreateImageStatus creates a new ImageStatus data model from a provided container. It updates a existing ImageStatus if provided
func CreateImageStatus(container pluginregistry.Container, attr ImageStatus) ImageStatus {
	containers := appendContainer(attr.Image.Containers, container)
	owners := appendOwner(attr.Image.MetaData.OwnerID, container.OwnerID)

	return ImageStatus{
		Image: Image{
			MetaData: pluginregistry.Image{
				ImageID: container.ImageID,
				Tag:     container.Image,
				OwnerID: owners,
			},
			Containers: containers,
		},
		State: Running,
	}
}
