package image

import (
	"testing"
	"time"

	"github.com/car2go/virity/internal/pluginregistry"
)

func TestAddToMonitored(t *testing.T) {

	monitored := NewMap()

	c1 := pluginregistry.Container{
		ID:        "testID1",
		Image:     "testImage1",
		ImageID:   "testImageID1",
		OwnerID:   "testOwner1",
		Timestamp: time.Now(),
	}

	c2 := pluginregistry.Container{
		ID:        "testID2",
		Image:     "testImage2",
		ImageID:   "testImageID2",
		OwnerID:   "testOwner1",
		Timestamp: time.Now(),
	}

	c3 := pluginregistry.Container{
		ID:        "testID3",
		Image:     "testImage1",
		ImageID:   "testImageID1",
		OwnerID:   "testOwner3",
		Timestamp: time.Now(),
	}

	c4 := pluginregistry.Container{
		ID:        "testID4",
		Image:     "testImage1",
		ImageID:   "testImageID1",
		OwnerID:   "testOwner1",
		Timestamp: time.Now(),
	}

	c5 := pluginregistry.Container{
		ID:        "testID5",
		Image:     "testImage2",
		ImageID:   "testImageID2",
		OwnerID:   "testOwner5",
		Timestamp: time.Now(),
	}

	c6 := pluginregistry.Container{
		ID:        "testID6",
		Image:     "testImage3",
		ImageID:   "testImageID3",
		OwnerID:   "testOwner1",
		Timestamp: time.Now(),
	}

	is1, _ := monitored.Read(c1.ImageID)
	monitored.Add(CreateImageStatus(c1, is1))
	is2, _ := monitored.Read(c2.ImageID)
	monitored.Add(CreateImageStatus(c2, is2))
	is3, _ := monitored.Read(c3.ImageID)
	monitored.Add(CreateImageStatus(c3, is3))
	is4, _ := monitored.Read(c4.ImageID)
	monitored.Add(CreateImageStatus(c4, is4))
	is5, _ := monitored.Read(c5.ImageID)
	monitored.Add(CreateImageStatus(c5, is5))
	is6, _ := monitored.Read(c6.ImageID)
	monitored.Add(CreateImageStatus(c6, is6))

	monitored.Range(func(k, v interface{}) bool {
		key := k.(string)
		val := v.(Data)
		t.Logf("Key: %v Containers: %v Owners: %v\n", key, val.Image.Containers, val.Image.MetaData.OwnerID)
		return true
	})

	i1, _ := monitored.Read("testImageID1")
	i2, _ := monitored.Read("testImageID2")
	i3, _ := monitored.Read("testImageID3")

	image1 := i1
	image2 := i2
	image3 := i3

	t.Log("\n\n")
	if image1.Image.Containers[0].ID != "testID1" {
		t.Errorf("wrong container ID found. Should be testID1 and is %v", image1.Image.Containers[0].ID)
	} else if image1.Image.Containers[1].ID != "testID3" {
		t.Errorf("wrong container ID found. Should be testID3 and is %v", image1.Image.Containers[1].ID)
	} else if image1.Image.Containers[2].ID != "testID4" {
		t.Errorf("wrong container ID found. Should be testID4 and is %v", image1.Image.Containers[2].ID)
	} else if image2.Image.Containers[0].ID != "testID2" {
		t.Errorf("wrong container ID found. Should be testID2 and is %v", image2.Image.Containers[0].ID)
	} else if image2.Image.Containers[1].ID != "testID5" {
		t.Errorf("wrong container ID found. Should be testID5 and is %v", image2.Image.Containers[1].ID)
	} else if image3.Image.Containers[0].ID != "testID6" {
		t.Errorf("wrong container ID found. Should be testID6 and is %v", image3.Image.Containers[0].ID)
	}

}

func TestEqualContainer(t *testing.T) {
	slice1 := []pluginregistry.Container{
		pluginregistry.Container{
			ID:        "testID1",
			Image:     "testImage1",
			ImageID:   "testImageID1",
			OwnerID:   "testOwner1",
			Timestamp: time.Now(),
		},
		pluginregistry.Container{
			ID:        "testID3",
			Image:     "testImage1",
			ImageID:   "testImageID1",
			OwnerID:   "testOwner3",
			Timestamp: time.Now(),
		},
		pluginregistry.Container{
			ID:        "testID2",
			Image:     "testImage2",
			ImageID:   "testImageID2",
			OwnerID:   "testOwner1",
			Timestamp: time.Now(),
		},
	}

	slice2 := []pluginregistry.Container{
		pluginregistry.Container{
			ID:        "testID1",
			Image:     "testImage1",
			ImageID:   "testImageID1",
			OwnerID:   "testOwner1",
			Timestamp: time.Now(),
		},
		pluginregistry.Container{
			ID:        "testID2",
			Image:     "testImage2",
			ImageID:   "testImageID2",
			OwnerID:   "testOwner1",
			Timestamp: time.Now(),
		},
		pluginregistry.Container{
			ID:        "testID3",
			Image:     "testImage1",
			ImageID:   "testImageID1",
			OwnerID:   "testOwner3",
			Timestamp: time.Now(),
		},
	}

	result := equalContainer(slice1, slice2)

	if !result {
		t.Errorf("%v, but should be true", result)
		return
	}

	slice1[1].ID = "testID4"

	result = equalContainer(slice1, slice2)

	if result {
		t.Errorf("%v, but should be false", result)
		return
	}
}
