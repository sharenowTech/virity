package schedule

import (
	"time"

	"github.com/car2go/virity/internal/log"
)

// Schedule a function using a interval
func Schedule(toDo func() error, interval time.Duration, name string) chan bool {
	ticker := time.NewTicker(interval)
	quit := make(chan bool)
	go func() {
		log.Info(log.Fields{
			"function": "Schedule",
			"package":  "schedule",
			"event":    name,
		}, "Starting Scheduler")
		// Execute once before waiting for ticks
		go toDo()
		for {
			select {
			case <-ticker.C:
				go toDo()
			case <-quit:
				log.Info(log.Fields{
					"function": "Schedule",
					"package":  "schedule",
					"event":    name,
				}, "Stopping Scheduler")
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}
