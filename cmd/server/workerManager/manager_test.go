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
