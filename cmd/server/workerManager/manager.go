package workerManager

import (
	"fmt"
	"sync"
	"time"

	"github.com/car2go/virity/cmd/server/worker"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

const errRetries = 5

// Manager contains all available agents, creates workers and assigns tasks
type Manager struct {
	agents sync.Map
}

// Plugins is a wrapper for worker plugins
type Plugins struct {
	Scanner pluginregistry.Scan
	Monitor pluginregistry.Monitor
	Store   pluginregistry.Store
}

type agent struct {
	data           pluginregistry.Agent
	latestCGroupID int64
	lastCheck      time.Time
	active         bool
}

var defManager = &Manager{}

// Delete agent from Scheduler
func (man *Manager) delete(key string) {
	man.agents.Delete(key)
}

// Refresh updates Manager and pull latest agents from store
func Refresh(store pluginregistry.Store) error {
	return defManager.Refresh(store)
}

// Refresh updates Manager and pull latest agents from store
func (man *Manager) Refresh(store pluginregistry.Store) error {
	agents, err := store.LoadAgents()
	if err != nil {
		return err
	}
	for _, elem := range agents {
		agentID := agentID(elem)
		if _, ok := man.agents.LoadOrStore(agentID, agent{
			data:   elem,
			active: true,
		}); !ok {
			log.Info(log.Fields{
				"package":  "main/workerManager",
				"function": "Refresh",
				"agent":    agentID,
			}, "New agent found")
		}
	}
	return nil
}

// Restore restores data from the store. It should be called only on first run
func Restore(p Plugins, cycleID int) error {
	return defManager.Restore(p, cycleID)
}

// Restore restores data from the store. It should be called only on first run
func (man *Manager) Restore(p Plugins, cycleID int) error {
	env := worker.Init(cycleID, p.Store, p.Scanner, p.Monitor)

	worker := env.InitMaintainWorker()
	err := worker.F.Restore(env)
	if err != nil {
		return err
	}
	log.Info(log.Fields{
		"package":  "main/workerManager",
		"function": "Restore",
	}, "Restored Monitored Data")
	return nil
}

// Run creates and manages new workers.
func Run(p Plugins, cycleID int) {
	defManager.Run(p, cycleID)
}

// Run creates and manages new workers.
func (man *Manager) Run(p Plugins, cycleID int) {
	var wg sync.WaitGroup

	env := worker.Init(cycleID, p.Store, p.Scanner, p.Monitor)

	man.agents.Range(func(k, v interface{}) bool {
		key := k.(string)
		val := v.(agent)

		cGroup, err := fetchCGroup(1*time.Minute, val.data, p.Store)
		if err != nil {
			log.Error(log.Fields{
				"package":  "main/workerManager",
				"function": "Run",
				"error":    err.Error(),
				"agent":    key,
			}, "Worker manager noticed an error while fetching a container group - this group/agent will be omitted in this cycle")
			return true
		}

		// If agent was just created and no lastestCGroupID exists
		if val.latestCGroupID != 0 {
			then := time.Unix(val.latestCGroupID, 0)
			duration := time.Since(then)

			if duration > val.data.Lifetime {
				val.active = false
				man.agents.Store(key, val)
				return true
			}
		}

		cWorker := env.InitCGroupWorker(*cGroup)

		if val.latestCGroupID == cGroup.ID {
			log.Debug(log.Fields{
				"package":       "main/workerManager",
				"function":      "Run",
				"agent":         key,
				"agent_active":  val.active,
				"last_cGroupID": val.latestCGroupID,
				"last_check":    val.lastCheck,
				"cGroupID":      cGroup.ID,
			}, "Run cGroupWorker for add active images")
			cWorker.Run(&wg, cWorker.F.Running)
		} else {
			log.Debug(log.Fields{
				"package":       "main/workerManager",
				"function":      "Run",
				"agent":         key,
				"agent_active":  val.active,
				"last_cGroupID": val.latestCGroupID,
				"last_check":    val.lastCheck,
				"cGroupID":      cGroup.ID,
			}, "Run cGroupWorker for analysing images")
			cWorker.Run(&wg, cWorker.F.Running, cWorker.F.Analyse)
		}

		val.lastCheck = time.Now()
		val.latestCGroupID = cGroup.ID
		man.agents.Store(key, val)

		return true
	})

	wg.Wait()

	mWorker := env.InitMaintainWorker()
	mWorker.Run(&wg, mWorker.F.Resolve, mWorker.F.Backup)
	return
}

// CleanUp store and manager data
func CleanUp(store pluginregistry.Store) error {
	return defManager.CleanUp(store)
}

// CleanUp store and manager data
func (man *Manager) CleanUp(store pluginregistry.Store) error {
	var globErr error
	man.agents.Range(func(k, v interface{}) bool {
		key := k.(string)
		val := v.(agent)

		if val.active == false {
			log.Debug(log.Fields{
				"package":       "main/workerManager",
				"function":      "CleanUp",
				"agent":         key,
				"agent_active":  val.active,
				"last_cGroupID": val.latestCGroupID,
				"last_check":    val.lastCheck,
			}, "Agent exceeded its lifetime and is still not active. It will be removed")
			man.delete(key)
			store.DeleteAgent(val.data)
			return true
		}
		err := store.Maintain(val.data, val.latestCGroupID)
		if err != nil {
			globErr = err
			return false
		}
		return true
	})

	if globErr != nil {
		return globErr
	}

	return nil
}

func fetchCGroup(timeout time.Duration, agent pluginregistry.Agent, store pluginregistry.Store) (*pluginregistry.ContainerGroup, error) {
	cGroup, err := store.LoadContainerGroup(agent)
	if err != nil {
		ticker := time.NewTicker(timeout / errRetries)
		chanCGroup := make(chan *pluginregistry.ContainerGroup)
		go func() {
			for t := range ticker.C {
				cGroup, err = store.LoadContainerGroup(agent)
				if err == nil {
					chanCGroup <- cGroup
					return
				}
				log.Warn(log.Fields{
					"package":  "main/workerManager",
					"function": "Run",
					"error":    err.Error(),
					"agent":    agentID(agent),
					"retry_at": t.Add(timeout / errRetries).Format("2006-01-02 15:04:05"),
				}, "Fetch Container Group failed with an error - I will try again")
			}
		}()
		select {
		case cGroup := <-chanCGroup:
			ticker.Stop()
			return cGroup, nil
		case <-time.After(timeout):
			ticker.Stop()
			time.Sleep(timeout / errRetries) // wait for last tick
			return nil, fmt.Errorf("Timeout error while fetching the Container Group")
		}
	}
	return cGroup, nil
}

func agentID(agent pluginregistry.Agent) string {
	return fmt.Sprintf("%v&%v&%v&%v", agent.Hostname, agent.DockerHostID, agent.Version, agent.RNG)
}
