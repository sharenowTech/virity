package kubernetes

import (
	"strings"

	"github.com/car2go/virity/cmd/agent/provider"
	"github.com/car2go/virity/internal/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// for interface validation
var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	provider.BaseProvider
}

// GetK8SPods uses the k8s client api to fetch all running pods
func getPods() ([]v1.Pod, error) {
	cli, err := connect()
	if err != nil {
		return nil, err
	}
	pods, err := cli.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	podList := pods.Items

	return podList, nil
}

func (p *Provider) extractContainers(pod v1.Pod) []provider.Container {
	list := make([]provider.Container, 0, 1)
	pod.GetLabels()
	for _, container := range pod.Status.ContainerStatuses {
		list = append(list, provider.Container{
			Name:     container.Name,
			ID:       validateHashes(container.ContainerID),
			Image:    container.Image,
			ImageID:  validateHashes(container.ImageID),
			Hostname: pod.Status.HostIP,
			OwnerID:  getOwner(pod, p.FallbackOwner, p.OwnerKey),
		})
	}
	return list
}

// GetRunningContainers returns a list of all containers in the kubernetes cluster
func (p Provider) GetRunningContainers() ([]provider.Container, error) {
	pods, err := getPods()
	if err != nil {
		return nil, err
	}

	list := make([]provider.Container, 0)

	for _, pod := range pods {
		for _, container := range p.extractContainers(pod) {
			list = append(list, container)
		}
	}
	return list, nil
}

// GetHostInfo returns Information of the Kubernetes Cluster
func (p Provider) GetHostInfo() (provider.HostInfo, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return provider.HostInfo{}, err
	}
	return provider.HostInfo{
		UUID:     "kubernetes:api",
		Hostname: config.Host,
	}, nil
}

func getOwner(pod v1.Pod, fallbackOwner string, ownerKey string) string {
	if owner, ok := pod.Labels[ownerKey]; ok {
		return owner
	}
	return fallbackOwner
}

func connect() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func validateHashes(hash string) string {

	index := strings.LastIndex(hash, ":")
	if index > -1 {
		log.Debug(log.Fields{
			"function": "validateHashes",
			"package":  "kubernetes",
			"hash":     hash,
		}, "removing sha265: prefix of image id")
		hash = hash[index+1:]
	}

	hash = strings.Replace(hash, "/", "", -1)

	return hash
}
