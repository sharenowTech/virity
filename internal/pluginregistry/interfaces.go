package pluginregistry

import (
	"time"
)

// MonitorStatus is the type for monitoring status based on an integer (to create a enum like structure)
type MonitorStatus int

const (
	StatusOK MonitorStatus = iota
	StatusWarning
	StatusError
)

// VulnSeverity is the type for Severities based on an integer (to create a enum like structure)
type VulnSeverity int

const (
	SeverityHigh VulnSeverity = iota
	SeverityMedium
	SeverityLow
	SeverityNegligible
)

// String implementes the String interface to VulnSeverity
func (s VulnSeverity) String() string {
	switch s {
	case SeverityHigh:
		return "High"
	case SeverityMedium:
		return "Medium"
	case SeverityLow:
		return "Low"
	case SeverityNegligible:
		return "Negligible"
	}
	return "not found"
}

// Scan is a interface for scan plugins like anchore or clair
type Scan interface {
	Scan(image Image) (*Vulnerabilities, error)
}

// Monitor is a interface for monitoring plugins like sensu
type Monitor interface {
	Push(image ImageStack, status MonitorStatus) error
	Resolve(image ImageStack) error
}

// Store is a interface for key value store plugins like etcd or redis
type Store interface {
	StoreContainer(container Container, agent Agent) error
	StoreImageStack(stack ImageStack, prefix string) error
	CheckID(agent Agent) (bool, error)
	LoadAgents() ([]Agent, error)
	LoadContainerGroup(agent Agent) (*ContainerGroup, error)
	LoadImageStacks(prefix string) ([]ImageStack, error)

	Maintain(agent Agent, latestStack int64) error // Function to keep store clean (remove old data etc.)

	DeleteAgent(agent Agent) error
	Delete(key string) error
}

// Config stores the configuration to initialize the plugin
type Config struct {
	Endpoint string
	User     string
	Password string

	CreateTickets   bool
	DefaultAssignee string //Asign tickets to when no user is defined default

}

// ContainerStack is the structure of multiple container sent in a specific time interval
type ContainerGroup struct {
	ID        int64
	Date      time.Time // Only for easier debugging
	Container []Container
}

// Container is the structure of a docker container
type Container struct {
	Name      string
	ID        string
	Hostname  string
	Image     string
	ImageID   string
	OwnerID   string
	Timestamp time.Time // required for stack id creation
}

// ImageStack is the combined struct of an Image, its vulnarbilities and the containers using this image
type ImageStack struct {
	MetaData   Image
	Vuln       Vulnerabilities
	Containers []Container //All containers which use the image
}

// Image is the structure of a docker image
type Image struct {
	Tag     string
	ImageID string
	StackID int64 // ID from which stack this image is extracted
	OwnerID []string
}

// Vulnerabilities container all found vulnerabilities of a image
type Vulnerabilities struct {
	Digest  string
	Scanner string
	CVE     []CVE
}

// CVE represents a single vulnerability
type CVE struct {
	Fix         string
	Package     string
	Severity    VulnSeverity
	URL         string
	Vuln        string
	Description string
}

type Agent struct {
	DockerHostID string
	Hostname     string
	Version      string
	RNG          string
	Lifetime     time.Duration
}
