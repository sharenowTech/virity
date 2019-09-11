package docker

import (
	"os"
	"strings"

	"github.com/sharenowTech/virity/cmd/agent/provider"
	"github.com/sharenowTech/virity/internal/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// for interface validation
var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	provider.BaseProvider
}

// GetRunningContainers uses the docker client api to fetch all running container on the docker host
func (p Provider) GetRunningContainers() ([]provider.Container, error) {
	cli, err := connectDocker()
	if err != nil {
		return nil, err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	cli.Close()

	list := make([]provider.Container, 0)

	for _, container := range containers {
		list = append(list, provider.Container{
			Name:     container.Names[0],
			ID:       container.ID,
			Hostname: getHostname(),
			Image:    container.Image,
			ImageID:  validateHashes(container.ImageID),
			OwnerID:  getDockerOwner(&container, p.FallbackOwner, p.OwnerKey),
		})
	}

	return list, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}
	return hostname
}

// GetHostInfo returns info of the current Docker Host
func (p Provider) GetHostInfo() (provider.HostInfo, error) {
	cli, err := connectDocker()
	if err != nil {
		return provider.HostInfo{}, err
	}
	info, err := cli.Info(context.Background())
	if err != nil {
		return provider.HostInfo{}, err
	}
	cli.Close()

	return provider.HostInfo{
		UUID:     info.ID,
		Hostname: getHostname(),
	}, nil
}

func getDockerOwner(container *types.Container, fallbackOwner string, ownerKey string) string {
	if owner, ok := container.Labels[ownerKey]; ok {
		return owner
	}
	return fallbackOwner
}

func connectDocker() (*client.Client, error) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient(client.DefaultDockerHost, "v1.32", nil, defaultHeaders)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func validateHashes(hash string) string {

	index := strings.LastIndex(hash, ":")
	if index > -1 {
		log.Debug(log.Fields{
			"function": "validateHashes",
			"package":  "docker",
			"hash":     hash,
		}, "removing sha265: prefix of image id")
		hash = hash[index+1:]
	}

	hash = strings.Replace(hash, "/", "", -1)

	return hash
}
