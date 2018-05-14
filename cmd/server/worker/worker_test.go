package worker

import (
	"testing"
	"time"

	"github.com/car2go/virity/cmd/server/image"
	"github.com/car2go/virity/internal/pluginregistry"
)

func TestRunning(t *testing.T) {

	date := time.Now()
	stack := pluginregistry.ContainerGroup{
		Container: []pluginregistry.Container{
			pluginregistry.Container{
				ID:        "foobar1",
				Hostname:  "fooServer1",
				Image:     "foo1",
				ImageID:   "bar1",
				Name:      "c1",
				Timestamp: date,
			},
			pluginregistry.Container{
				ID:        "foobar2",
				Hostname:  "fooServer2",
				Image:     "foo2",
				ImageID:   "bar2",
				Name:      "c2",
				Timestamp: date,
			},
			pluginregistry.Container{
				ID:        "foobar3",
				Hostname:  "fooServer3",
				Image:     "foo1",
				ImageID:   "bar1",
				Name:      "c3",
				Timestamp: date,
			},
		},
	}

	env := Environment{
		RunningImages: &image.Active{},
	}
	//var wg sync.WaitGroup

	stackW := env.InitCGroupWorker(stack)
	//maintainW := env.InitMaintainWorker

	stackW.F.Running(env, stack.Container[0])
	stackW.F.Running(env, stack.Container[1])
	stackW.F.Running(env, stack.Container[2])

	_, err := env.RunningImages.Get(stack.Container[0].ImageID)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = env.RunningImages.Get(stack.Container[1].ImageID)
	if err != nil {
		t.Error(err)
		return
	}
}
