package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sharenowTech/virity/cmd/agent/provider"
	"github.com/sharenowTech/virity/cmd/agent/provider/docker"
	"github.com/sharenowTech/virity/cmd/agent/provider/kubernetes"

	"github.com/sharenowTech/virity/internal/config"
	"github.com/sharenowTech/virity/internal/log"
	"github.com/sharenowTech/virity/internal/pluginregistry"
	"github.com/sharenowTech/virity/internal/schedule"
	_ "github.com/sharenowTech/virity/internal/store/etcd"
)

const ownerKey = "virity.owner"

var version string
var Environment = config.GetGeneralConfig().AgentEnv

type Version struct {
	Major string
	Minor string
	Build string
}

func decodeVersion(s string) Version {
	split := strings.Split(s, ".")
	if len(split) == 3 {
		return Version{
			Major: split[0],
			Minor: split[1],
			Build: split[2],
		}
	}
	return Version{
		Major: s,
		Minor: s,
		Build: s,
	}
}

func generateID(seed int64) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(seed)
	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func GenerateAgentData(hostname string, dockerHostID string) (pluginregistry.Agent, error) {
	gencfg := config.GetGeneralConfig()
	id := generateID(gencfg.AgentSeed)
	cfg := config.GetStoreConfig()
	store, err := pluginregistry.NewStore(cfg.Type, pluginregistry.Config{
		Endpoint: cfg.Endpoint,
	})
	if err != nil {
		return pluginregistry.Agent{}, err
	}
	agent := pluginregistry.Agent{
		DockerHostID: dockerHostID,
		Version:      version,
		Hostname:     hostname,
		RNG:          id,
		Lifetime:     gencfg.AgentLifetime,
	}
	if ok, _ := checkAgent(store, agent); !ok {
		return agent, fmt.Errorf("AgentID already exists - existing data will be overwritten")
	}
	return agent, nil
}

func checkAgent(store pluginregistry.Store, id pluginregistry.Agent) (bool, error) {
	return store.CheckID(id)
}

func sendContainer(store pluginregistry.Store, host provider.HostInfo, list []provider.Container) {
	log.Info(log.Fields{
		"function": "sendContainer",
		"package":  "main",
		"count":    len(list),
	}, "store containers")

	var wg sync.WaitGroup

	timestamp := time.Now()
	for _, container := range list {

		container.Timestamp = timestamp

		agent, err := GenerateAgentData(host.Hostname, host.UUID)
		if err != nil {
			log.Info(log.Fields{
				"function": "GenerateAgentData",
				"package":  "main",
				"error":    err.Error(),
				"hostname": agent.Hostname,
				"dockerID": agent.DockerHostID,
				"version":  version,
			}, "An error occoured generating an AgentID")
		}

		wg.Add(1)
		// make push faster
		go func(container provider.Container, wg *sync.WaitGroup) {
			defer wg.Done()
			storeErr := store.StoreContainer(container.Convert(), agent)
			if storeErr != nil {
				log.Error(log.Fields{
					"function": "sendContainer",
					"package":  "main",
					"error":    storeErr.Error(),
				}, "could not store container")
				return
			}
		}(container, &wg)
	}
	wg.Wait()
}

func Send() error {
	storeConfig := config.GetStoreConfig()
	store, err := pluginregistry.NewStore(storeConfig.Type, pluginregistry.Config{
		Endpoint: storeConfig.Endpoint,
	})
	if err != nil {
		log.Error(log.Fields{
			"function": "Send",
			"package":  "main",
			"error":    err.Error(),
		}, "could not create store")
		return err
	}

	var base provider.Provider
	baseConfig := provider.BaseProvider{
		OwnerKey:      ownerKey,
		FallbackOwner: config.GetMonitorConfig().DefaultAssignee,
	}

	switch Environment {
	case "k8s":
		base = kubernetes.Provider{BaseProvider: baseConfig}
	default:
		base = docker.Provider{BaseProvider: baseConfig}
	}
	host, containerList, err := getContainers(base)
	if err != nil {
		log.Error(log.Fields{
			"function": "Send",
			"package":  "main",
			"error":    err.Error(),
		}, "could not fetch container information")
		return err
	}

	sendContainer(store, host, containerList)

	return nil
}

func getContainers(p provider.Provider) (provider.HostInfo, []provider.Container, error) {
	containerList, err := p.GetRunningContainers()
	if err != nil {
		return provider.HostInfo{}, nil, err
	}
	host, err := p.GetHostInfo()
	if err != nil {
		return provider.HostInfo{}, nil, err
	}
	return host, containerList, nil
}

func main() {
	v := decodeVersion(version)
	log.Info(log.Fields{
		"major":    v.Major,
		"minor":    v.Minor,
		"build":    v.Build,
		"function": "main",
		"package":  "main",
		"version":  version,
	}, "Version")
	conf := config.GetStoreConfig()

	if Environment == "batch" {
		fmt.Println("Batchflag is set")
		err := Send()
		if err != nil {
			log.Critical(log.Fields{
				"function": "main",
				"package":  "main",
				"error":    err.Error(),
			}, "Agent could not be started")
			os.Exit(2)
		}
		return
	}
	quit := schedule.Schedule(Send, conf.IntervalAgentPush, "Push")
	<-quit
}
