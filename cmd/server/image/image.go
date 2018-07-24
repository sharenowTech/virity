package image

import (
	"sort"

	"github.com/car2go/virity/internal/pluginregistry"
)

const backupPath = "Backup/Monitored"

func equalContainer(slice1, slice2 []pluginregistry.Container) bool {

	if len(slice1) != len(slice2) {
		return false
	}

	if (slice1 == nil) != (slice2 == nil) {
		return false
	}

	sort.Slice(slice1, func(i, j int) bool { return slice1[i].ID < slice1[j].ID })
	sort.Slice(slice2, func(i, j int) bool { return slice2[i].ID < slice2[j].ID })

	slice2 = slice2[:len(slice1)]
	for index, val := range slice1 {
		if val.ID != slice2[index].ID {
			return false
		}
	}
	return true
}

// Difference returns the elements in slice1 that aren't in slice2
func difference(slice1, slice2 []string) []string {
	mapSlice2 := map[string]bool{}
	for _, elem := range slice2 {
		mapSlice2[elem] = true
	}
	diff := []string{}
	for _, elem := range slice1 {
		if _, ok := mapSlice2[elem]; !ok {
			diff = append(diff, elem)
		}
	}
	return diff
}
