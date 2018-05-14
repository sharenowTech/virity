package etcd

import (
	"testing"
	"time"

	"github.com/car2go/virity/internal/config"

	"github.com/car2go/virity/internal/pluginregistry"
)

func TestStore(t *testing.T) {
	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}
	//date, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Jan 2, 2018 at 3:04pm (MST)")
	//t.Log(err)
	timestamp := time.Now()
	data := pluginregistry.Container{
		ID:        "foobar",
		Hostname:  "fooServer",
		Name:      "foo",
		Image:     "foo",
		ImageID:   "bar",
		Timestamp: timestamp,
	}
	err := etcd.StoreContainer(data, pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})
	if err != nil {
		t.Error(err.Error())
	}

	data2 := pluginregistry.Container{
		ID:        "foobar2",
		Hostname:  "fooServer2",
		Image:     "foo2",
		Name:      "foo",
		ImageID:   "bar2",
		Timestamp: timestamp,
	}
	err = etcd.StoreContainer(data2, pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestLoadStack(t *testing.T) {
	time.Sleep(2 * time.Second)
	TestStore(t)
	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}
	//date, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Jan 2, 2018 at 3:04pm (MST)")
	//t.Log(err)
	resp, err := etcd.LoadContainerGroup(pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	})

	if err != nil {
		t.Error(err.Error())
	}
	if resp.Container[0].ID != "foobar" {
		t.Error("1: Wrong ID:" + resp.Container[0].ID)
	} else if resp.Container[0].Hostname != "fooServer" {
		t.Error("1: Wrong Hostname" + resp.Container[0].Hostname)
	} else if resp.Container[0].Image != "foo" {
		t.Error("1: Wrong Image Name" + resp.Container[0].Image)
	} else if resp.Container[0].ImageID != "bar" {
		t.Error("1: Wrong Image ID" + resp.Container[0].ImageID)
	} else if resp.Container[1].ID != "foobar2" {
		t.Error("2: Wrong ID" + resp.Container[1].ID)
	} else if resp.Container[1].Hostname != "fooServer2" {
		t.Error("2: Wrong Hostname" + resp.Container[1].Hostname)
	} else if resp.Container[1].Image != "foo2" {
		t.Error("2: Wrong Image Name" + resp.Container[1].Image)
	} else if resp.Container[1].ImageID != "bar2" {
		t.Error("2: Wrong Image ID" + resp.Container[1].ImageID)
	}
}

func TestMaintain(t *testing.T) {

	TestStore(t)
	time.Sleep(1 * time.Second)

	currTime := time.Now()
	TestStore(t)

	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}
	api, err := etcd.connect()
	if err != nil {
		t.Error(err)
		return
	}

	latestKey, err := latestStackKey(api, encodeAgentID(pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	}, agentsPath))
	if err != nil {
		t.Error(err)
		return
	}

	err = etcd.Maintain(pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	}, latestKey)
	if err != nil {
		t.Error(err)
		return
	}

	latestKey, err = latestStackKey(api, encodeAgentID(pluginregistry.Agent{
		DockerHostID: "agent1",
		Hostname:     "myhost",
		RNG:          "0",
		Version:      "latest",
		Lifetime:     2 * time.Hour,
	}, agentsPath))
	if err != nil {
		t.Error(err)
		return
	}

	if latestKey != currTime.Unix() {
		t.Errorf("Wrong Key was kept in the store after maintain")
		return
	}
}

func TestListAgents(t *testing.T) {
	TestStore(t)
	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}

	agentsPath, err := etcd.LoadAgents()
	if err != nil {
		t.Error(err)
	}

	for _, agent := range agentsPath {
		if agent.DockerHostID == "agent1" && agent.Lifetime == 2*time.Hour {
			return
		}
	}
}

func TestStoreImageStack(t *testing.T) {
	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}
	timestamp := time.Now()
	img := pluginregistry.ImageStack{
		MetaData: pluginregistry.Image{
			ImageID: "myImageID",
			Tag:     "myImageTag",
			OwnerID: []string{"owner1", "owner2", "owner3"},
		},
		Vuln: pluginregistry.Vulnerabilities{
			Scanner: "myScanner",
			Digest:  "myDigest",
			CVE: []pluginregistry.CVE{
				pluginregistry.CVE{
					Fix:         "myFix1",
					Package:     "myPackage1",
					Severity:    pluginregistry.SeverityHigh,
					URL:         "myURL1",
					Vuln:        "myVuln1",
					Description: "myDesc1",
				},
			},
		},
		Containers: []pluginregistry.Container{
			pluginregistry.Container{
				Name:      "myName2",
				ID:        "myID2",
				Hostname:  "myHostname2",
				Image:     "myImage2",
				ImageID:   "myImageID2",
				OwnerID:   "owner2",
				Timestamp: timestamp,
			},
			pluginregistry.Container{
				Name:      "myName3",
				ID:        "myID3",
				Hostname:  "myHostname3",
				Image:     "myImage3",
				ImageID:   "myImageID3",
				OwnerID:   "owner3",
				Timestamp: timestamp,
			},
			pluginregistry.Container{
				Name:      "myName1",
				ID:        "myID1",
				Hostname:  "myHostname1",
				Image:     "myImage1",
				ImageID:   "myImageID1",
				OwnerID:   "owner1",
				Timestamp: timestamp,
			},
		},
	}
	if err := etcd.StoreImageStack(img, "test/images"); err != nil {
		t.Error(err)
		return
	}
}

func TestLoadImageStack(t *testing.T) {
	time.Sleep(2 * time.Second)
	TestStoreImageStack(t)
	storecfg := config.GetStoreConfig()
	etcd := etcd{
		endpoint: storecfg.Endpoint,
	}
	//date, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Jan 2, 2018 at 3:04pm (MST)")
	//t.Log(err)
	resp, err := etcd.LoadImageStacks("test/images")
	if err != nil {
		t.Error(err)
		return
	}

	for _, img := range resp {
		t.Log(img)
		if img.MetaData.ImageID != "myImageID" {
			t.Error("ID parsing failed")
			return
		} else if img.MetaData.Tag != "myImageTag" {
			t.Error("Tag parsing failed")
			return
		} else if img.Vuln.CVE[0].Vuln != "myVuln1" {
			t.Error("CVE parsing failed")
			return
		}
	}
}
