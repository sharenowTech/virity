package etcd

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sharenowTech/virity/internal/pluginregistry"
	"github.com/coreos/etcd/client"
)

func TestBucketBaseKey(t *testing.T) {

	if bucketBaseKey(time.Unix(1513953420, 0), 30*time.Minute) != "1513953000" {
		t.Error("time should be 1513953000")
	}

	if bucketBaseKey(time.Unix(1513953408, 0), 17*time.Second) != "1513953394" {
		t.Error("time should be 1513953394")
	}

	if bucketBaseKey(time.Unix(1513954980, 0), 2*time.Hour) != "1513951200" {
		t.Error("time should be 1513951200")
	}

	if bucketBaseKey(time.Unix(0, 0), 2*time.Hour) != "0" {
		t.Error("time should be 1513951200")
	}
}

func TestCreateContainer(t *testing.T) {

	node := client.Node{
		Key: "foobar",
		Nodes: client.Nodes{
			&client.Node{
				Key:   "Hostname",
				Value: "fooServer",
			}, &client.Node{
				Key:   "ID",
				Value: "foobar",
			}, &client.Node{
				Key:   "Image",
				Value: "foo",
			}, &client.Node{
				Key:   "ImageID",
				Value: "bar",
			}, &client.Node{
				Key:   "Timestamp",
				Value: "1521627056",
			},
		},
	}

	container, err := parseContainer(&node)
	if err != nil {
		t.Error(err)
		return
	}
	//parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02 15:04:05.999999999 -0700 MST")
	if container.ID != "foobar" {
		t.Error("ID parsing failed")
	} else if container.Hostname != "fooServer" {
		t.Error("Hostname parsing failed")
	} else if container.Image != "foo" {
		t.Error("Image parsing failed")
	} else if container.ImageID != "bar" {
		t.Error("ImageID parsing failed")
	} else if container.Timestamp.String() != "2018-03-21 10:10:56 +0000 UTC" {
		t.Log("Container: " + container.Timestamp.String())
		t.Error("Timestamp parsing failed")
	}
}

func TestCreateNode(t *testing.T) {

	date, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02 15:04:05.999999999 -0700 MST")
	container := pluginregistry.Container{
		ID:        "foobar",
		Hostname:  "fooServer",
		Image:     "foo",
		ImageID:   "bar",
		Timestamp: date,
	}

	node, err := createNode(reflect.ValueOf(container), container.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if node.Key != "foobar" {
		t.Error("ID parsing failed")
	}
	for _, value := range node.Nodes {
		switch value.Key {
		case hostname:
			if value.Value != "fooServer" {
				t.Error("Hostname parsing failed")
			}
		case imageID:
			if value.Value != "bar" {
				t.Error("Hostname parsing failed")
			}
		case image:
			if value.Value != "foo" {
				t.Error("Hostname parsing failed")
			}
		case timestamp:
			container.Timestamp, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", value.Value)
			if value.Value != "2006-01-02 15:04:05.999999999 -0700 MST" {
				t.Error("Hostname parsing failed")
			}
		default:
		}
	}
}

func TestCreateField(t *testing.T) {
	img := pluginregistry.ImageStack{
		MetaData: pluginregistry.Image{
			ImageID: "myImageID",
			Tag:     "myImageTag",
		},
		Vuln: pluginregistry.Vulnerabilities{
			Scanner: "myScanner",
			Digest:  "myDigest",
			CVE: []pluginregistry.CVE{
				pluginregistry.CVE{},
				pluginregistry.CVE{},
				pluginregistry.CVE{},
			},
		},
		Containers: []pluginregistry.Container{
			pluginregistry.Container{},
			pluginregistry.Container{},
			pluginregistry.Container{},
		},
	}

	foo := reflect.ValueOf(img)

	node, err := createNode(foo, "struct")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%+v", node.Nodes[2].Nodes)

}
