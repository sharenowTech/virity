package provider

import "github.com/sharenowTech/virity/internal/pluginregistry"

type Provider interface {
	GetHostInfo() (HostInfo, error)
	GetRunningContainers() ([]Container, error)
}

type BaseProvider struct {
	FallbackOwner string
	OwnerKey      string
}

type HostInfo struct {
	UUID     string
	Hostname string
}

type Container pluginregistry.Container

func (c Container) Convert() pluginregistry.Container {
	return pluginregistry.Container(c)
}
