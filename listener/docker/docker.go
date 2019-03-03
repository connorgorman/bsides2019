package docker

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	demoTypes "github.com/connorgorman/bsides2019/types"
	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Listener struct {
	client                 *client.Client
	newContainerChannel    chan demoTypes.Container
	removeContainerChannel chan string
}

func NewListener() (*Listener, error) {
	possibleDockerSockets := []string{
		"unix:///host/var/run/docker.sock",
		"unix:///host/run/docker.sock",
	}

	var (
		dockerClient *client.Client
		err          error
	)
	for _, s := range possibleDockerSockets {
		dockerClient, err = client.NewClient(s, api.DefaultVersion, nil, nil)
		if err != nil {
			return nil, err
		}
		_, err = dockerClient.Info(context.Background())
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &Listener{
		client:                 dockerClient,
		newContainerChannel:    make(chan demoTypes.Container),
		removeContainerChannel: make(chan string),
	}, nil
}

func (d *Listener) NewContainerChannel() <-chan demoTypes.Container {
	return d.newContainerChannel
}

func (d *Listener) RemoveContainerChannel() <-chan string {
	return d.removeContainerChannel
}

func (d *Listener) inspectContainerAndPush(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerJSON, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}
	if containerJSON.ContainerJSONBase == nil {
		return fmt.Errorf("could not find ContainerJSONBase and therefore cannot pull GraphDriver data")
	}
	path, ok := containerJSON.ContainerJSONBase.GraphDriver.Data["MergedDir"]
	if !ok {
		return fmt.Errorf("could not find MergedDir for containerJSON %q", id)
	}
	d.newContainerChannel <- demoTypes.Container{
		ID:   containerJSON.ID,
		Name: containerJSON.Config.Labels["io.kubernetes.container.name"],
		Pod:  containerJSON.Config.Labels["io.kubernetes.pod.name"],

		PID:      containerJSON.State.Pid,
		FilePath: filepath.Join("/host/", path),
	}
	return nil
}

func (d *Listener) events() {
	eventFilters := filters.NewArgs()
	eventFilters.Add("type", "container")
	eventFilters.Add("event", "start")
	eventFilters.Add("event", "stop")
	eventFilters.Add("event", "kill")

	eventChan, errorChan := d.client.Events(context.Background(), types.EventsOptions{Filters: eventFilters})
	for {
		select {
		case event := <-eventChan:
			switch event.Action {
			case "start":
				if err := d.inspectContainerAndPush(event.Actor.ID); err != nil {
					log.Printf("error handling container id %q", event.Actor.ID)
				}
			case "stop", "kill":
				d.removeContainerChannel <- event.Actor.ID
				log.Printf("KILL or STOP: %q", event.Actor.ID)
			}
		case err := <-errorChan:
			log.Printf("Error: %v", err)
		}
	}
}

func (d *Listener) Start() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
	}
	for _, container := range containers {
		if err := d.inspectContainerAndPush(container.ID); err != nil {
			log.Println(err)
		}
	}

	d.events()
}
