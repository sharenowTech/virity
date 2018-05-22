package workerManager

import (
	"testing"
	"time"

	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/pluginregistry"
	_ "github.com/car2go/virity/internal/store/etcd"
)

func TestManagerUpdate(t *testing.T) {

	configStore := config.GetStoreConfig()
	store, err := pluginregistry.NewStore(configStore.Type, pluginregistry.Config{
		Endpoint: configStore.Endpoint,
	})
	if err != nil {
		t.Error(err)
		return
	}

	store.StoreContainer(pluginregistry.Container{
		ID:        "testID",
		Image:     "testImage",
		ImageID:   "testImageID",
		OwnerID:   "testOwner",
		Timestamp: time.Now(),
	}, pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})

	store.StoreContainer(pluginregistry.Container{
		ID:        "testID2",
		Image:     "testImage2",
		ImageID:   "testImageID2",
		OwnerID:   "testOwner2",
		Timestamp: time.Now(),
	}, pluginregistry.Agent{
		DockerHostID: "agent2",
		Hostname:     "myhost",
		RNG:          "1",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})

	store.StoreContainer(pluginregistry.Container{
		ID:        "testID2",
		Image:     "testImage2",
		ImageID:   "testImageID2",
		OwnerID:   "testOwner2",
		Timestamp: time.Now(),
	}, pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})

	err = defManager.Refresh(store)
	if err != nil {
		t.Error(err)
		return
	}

	_, ok := defManager.agents.Load(agentID(pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	}))
	if !ok {
		t.Errorf("agent1 not found")
	}
	_, ok = defManager.agents.Load(agentID(pluginregistry.Agent{
		DockerHostID: "agent2",
		Hostname:     "myhost",
		RNG:          "1",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	}))
	if !ok {
		t.Errorf("agent2 not found")
	}
}

func TestSplitCGroup(t *testing.T) {

	g := pluginregistry.ContainerGroup{
		ID: 123,
		Container: []pluginregistry.Container{
			pluginregistry.Container{
				ID: "1",
			},
			pluginregistry.Container{
				ID: "2",
			},
			pluginregistry.Container{
				ID: "3",
			},
			pluginregistry.Container{
				ID: "4",
			},
			pluginregistry.Container{
				ID: "5",
			},
			pluginregistry.Container{
				ID: "6",
			},
			pluginregistry.Container{
				ID: "7",
			},
			pluginregistry.Container{
				ID: "8",
			},
			pluginregistry.Container{
				ID: "9",
			},
			pluginregistry.Container{
				ID: "10",
			},
			pluginregistry.Container{
				ID: "11",
			},
			pluginregistry.Container{
				ID: "12",
			},
			pluginregistry.Container{
				ID: "13",
			},
			pluginregistry.Container{
				ID: "14",
			},
			pluginregistry.Container{
				ID: "15",
			},
			pluginregistry.Container{
				ID: "16",
			},
			pluginregistry.Container{
				ID: "17",
			},
		},
	}

	groups := splitCGroup(&g, 6)

	for _, slice := range groups {
		t.Log(slice)
	}

	if len(groups[0].Container) != 3 {
		t.Error("Length did not match")
		return
	}

	if len(groups[5].Container) != 2 {
		t.Error("Length did not match")
		return
	}

}
