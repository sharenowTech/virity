package image

import (
	"fmt"
	"sort"

	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

const backupPath = "Backup/Monitored"

// Action defines an enum type to tag images with a "todo"
// e.g. FullyResolve to remove the image if it is not running anymore
type Action int

// Status defines an enum type to tag images with their current status
// e.g. Scanning if an image is currently being scanned
type Status int

const (
	// Noop (No Operation) is default.
	noop Action = iota
	// Update image --> something has changed (new container etc.)
	Update
	// PartlyResolve image --> some containers with this image are not running anymore. Remove these containers and owner from the image
	PartlyResolve
	// FullyResolve image --> the image is not running anymore. It will be removed completely
	FullyResolve
)

const (
	// Running --> Image is running but not yet monitored or scanned
	Running Status = iota
	// Scanning --> Image is currently being scanned
	Scanning
	// Scanned --> Image is scanned but not yet monitored
	Scanned
	// Monitored --> Image data is sent to a monitoring tool
	Monitored
	// Resolved --> Image is not running anymore and therefore is is resolved
	Resolved
)

// ImageStatus contains the image, the current state of the image (running, scanned or monitored) and if it needs to be resolved or updated
type ImageStatus struct {
	Image          Image
	State          Status
	Action         Action
	StateChangedAt int
}

// Image is a wrapper for the pluginregistry ImageStack
type Image pluginregistry.ImageStack

// Model is an interface of the underlying model
type Model interface {
	Add(image ImageStatus)
	Delete(image ImageStatus)
	Read(id string) (val ImageStatus, ok bool)
	Range(f func(key, val interface{}) bool)
	Reset()
	UpdateState(state Status, cycleID int, attr ImageStatus) ImageStatus
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
