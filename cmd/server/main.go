package main

import (
	"strings"

	"github.com/car2go/virity/cmd/server/workerManager"
	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/log"
	_ "github.com/car2go/virity/internal/monitoring/sensu"
	_ "github.com/car2go/virity/internal/scanner/anchore"
	"github.com/car2go/virity/internal/schedule"
	_ "github.com/car2go/virity/internal/store/etcd"
)

type imageID string

var version string
var cycle int

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

func scheduled() error {
	cycle++
	log.Debug(log.Fields{
		"package":  "main",
		"function": "scheduled",
		"cycleID":  cycle,
	}, "------------ Starting new cycle ------------")
	store, err := createStore()
	scan, err := createScanner()
	monitor, err := createMonitor()
	if err != nil {
		log.Error(log.Fields{
			"package":  "main",
			"function": "scheduled",
			"error":    err.Error(),
		}, "An error occourd while getting plugins")
		return err
	}
	p := workerManager.Plugins{
		Store:   store,
		Scanner: scan,
		Monitor: monitor,
	}

	if cycle == 1 {
		err = workerManager.Restore(p, cycle)
	}
	if err != nil {
		log.Error(log.Fields{
			"package":  "main",
			"function": "scheduled",
			"error":    err.Error(),
		}, "An error occurred while restoring data")
		//continue with clean new structure
	}

	err = workerManager.Refresh(p.Store)
	if err != nil {
		log.Error(log.Fields{
			"package":  "main",
			"function": "scheduled",
			"error":    err.Error(),
		}, "An error occurred while refreshing agents in worker manager")
		return err
	}

	workerManager.Run(p, cycle)

	err = workerManager.CleanUp(p.Store)
	if err != nil {
		log.Error(log.Fields{
			"package":  "main",
			"function": "scheduled",
			"error":    err.Error(),
		}, "An error occourd while clean up agents in worker manager")
		return err
	}
	log.Debug(log.Fields{
		"package":  "main",
		"function": "scheduled",
		"cycleID":  cycle,
	}, "------------ End of cycle ------------")
	return nil
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
	config := config.GetStoreConfig()

	quit := schedule.Schedule(scheduled, config.IntervalServerPull, "Scan/Monitor")

	switch {
	case <-quit:
		//maintainQuit <- true
	}

}
