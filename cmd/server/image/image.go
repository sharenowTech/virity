package image

import (
	"fmt"
	"sort"
	"sync"

	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
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
