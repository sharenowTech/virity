package workerManager

import (
	"fmt"
	"sync"
	"time"

	"github.com/car2go/virity/cmd/server/worker"
	"github.com/car2go/virity/cmd/server/worker/task"
	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
)

const workers = 10
const errRetries = 5

// Manager contains all available agents, creates workers and assigns tasks
type Manager struct {
	agents     sync.Map
	dispatcher *worker.Dispatcher
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

func init() {
	defManager.dispatcher = worker.NewDispatcher(workers)
	defManager.dispatcher.Run()
}

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
	var wg sync.WaitGroup

	template := task.New(&wg, cycleID, 0, p.Store, p.Scanner, p.Monitor)
	t := template.Maintain(task.Restore)
	task.AddToQueue(&t)

	wg.Wait()

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
	running := pluginregistry.ContainerGroup{}
	analyse := pluginregistry.ContainerGroup{}

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

		// If agent was not just created (if agent has a lastestCGroupID)
		if val.latestCGroupID != 0 {
			then := time.Unix(val.latestCGroupID, 0)
			duration := time.Since(then)

			// if agent is inactive
			if duration > val.data.Lifetime {
				val.active = false
				man.agents.Store(key, val)
				return true
			}
		}

		// If something has changed during last fetch, analyze these images
		if val.latestCGroupID != cGroup.ID {
			analyse.Container = append(analyse.Container, cGroup.Container...)
		}
		running.Container = append(running.Container, cGroup.Container...)

		val.lastCheck = time.Now()
		val.latestCGroupID = cGroup.ID
		man.agents.Store(key, val)

		return true
	})

	log.Debug(log.Fields{
		"package":  "main/workerManager",
		"function": "Run",
		"count":    len(running.Container),
	}, "Fetched Containers")

	var wg sync.WaitGroup

	template := task.New(&wg, cycleID, 5, p.Store, p.Scanner, p.Monitor)

	for _, container := range running.Container {
		t := template.Container(container, task.Running)
		task.AddToQueue(&t)
	}

	for _, container := range analyse.Container {
		t := template.Container(container, task.Analyse)
		task.AddToQueue(&t)
	}

	wg.Wait()

	template.Retries = 0

	t := template.Maintain(task.Resolve, task.Backup)
	task.AddToQueue(&t)

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
