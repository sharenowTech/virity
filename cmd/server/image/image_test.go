package image

import (
	"testing"
	"time"

	"github.com/car2go/virity/internal/pluginregistry"
)

func TestAddToMonitored(t *testing.T) {

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

	defMonitored.add(c1)
	defMonitored.add(c2)
	defMonitored.add(c3)
	defMonitored.add(c4)
	defMonitored.add(c5)
	defMonitored.add(c6)

	defMonitored.images.Range(func(k, v interface{}) bool {
		key := k.(string)
		val := v.(imageStatus)
		t.Logf("Key: %v Containers: %v Owners: %v\n", key, val.image.Containers, val.image.MetaData.OwnerID)
		return true
	})

	i1, _ := defMonitored.images.Load("testImageID1")
	i2, _ := defMonitored.images.Load("testImageID2")
	i3, _ := defMonitored.images.Load("testImageID3")

	image1 := i1.(imageStatus)
	image2 := i2.(imageStatus)
	image3 := i3.(imageStatus)

	t.Log("\n\n")
	if image1.image.Containers[0].ID != "testID1" {
		t.Errorf("wrong container ID found. Should be testID1")
	} else if image1.image.Containers[1].ID != "testID3" {
		t.Errorf("wrong container ID found. Should be testID3")
	} else if image1.image.Containers[2].ID != "testID4" {
		t.Errorf("wrong container ID found. Should be testID4")
	} else if image2.image.Containers[0].ID != "testID2" {
		t.Errorf("wrong container ID found. Should be testID2")
	} else if image2.image.Containers[1].ID != "testID5" {
		t.Errorf("wrong container ID found. Should be testID5")
	} else if image3.image.Containers[0].ID != "testID6" {
		t.Errorf("wrong container ID found. Should be testID6")
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
