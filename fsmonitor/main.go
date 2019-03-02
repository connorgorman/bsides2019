package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fsnotify/fsnotify"
)

var pathsToIgnore = map[string]struct{}{
	"proc":{},
	"dev":{},
}

var pathLength = 0

type Container struct {
	ID, Name, Pod string
	PID int

	modifiedPaths map[string]struct{}
	listeningPaths map[string]struct{}
}

type ContainerManager struct {
	client *client.Client
	watcher *fsnotify.Watcher

	containerIDToContainer map[string]*Container
	mergedPathToContainer map[string]*Container
}

func (c *ContainerManager) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}
	for _, container := range containers {
		if err := c.getContainerAndAddToWatcher(container.ID); err != nil {
			log.Printf("Error adding container %q: %v", container.ID, err)
		}
	}
	return nil
}

func (c *ContainerManager) AddToWatcher(container *Container, files ...string) {
	for _, f := range files {
		if err := c.watcher.Add(f); err != nil {
			log.Printf("error adding file %q to watcher", f)
			continue
		}
		container.listeningPaths[f] = struct{}{}
	}
}

func getAllSubDirectories(dir string) []string {
	var dirs []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if _, ok := pathsToIgnore[info.Name()]; ok {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		log.Printf("Error getting subdirectories for subdir: %v", err)
	}
	return dirs
}

func (c *ContainerManager) getContainerAndAddToWatcher(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerJSON, err := c.client.ContainerInspect(ctx, id)
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

	filesystemPathInContainer := filepath.Join("/host/", path)
	container := &Container{
		ID: containerJSON.ID,
		Name: containerJSON.Config.Labels["io.kubernetes.container.name"],
		Pod: containerJSON.Config.Labels["io.kubernetes.pod.name"],

		PID: containerJSON.State.Pid,

		modifiedPaths: make(map[string]struct{}),
		listeningPaths: make(map[string]struct{}),
	}
	c.containerIDToContainer[container.ID] = container

	// Set the index of the prefix for the container filesystems
	// this allows for quick indexing when getting the filepaths within the container
	if pathLength == 0 {
		pathLength = len(filesystemPathInContainer)
	}

	c.mergedPathToContainer[filesystemPathInContainer] = container
	dirs := getAllSubDirectories(filesystemPathInContainer)

	c.AddToWatcher(container, dirs...)
	return nil
}

func (c *ContainerManager) WatchDockerEvents() {
	eventFilters := filters.NewArgs()
	eventFilters.Add("type", "container")
	eventFilters.Add("event", "start")
	eventFilters.Add("event", "stop")
	eventFilters.Add("event", "kill")

	eventChan, errorChan := c.client.Events(context.Background(), types.EventsOptions{Filters: eventFilters})
	for {
		select {
		case event := <-eventChan:
			switch event.Action {
			case "start":
				if err := c.getContainerAndAddToWatcher(event.Actor.ID); err != nil {
					log.Printf("error: %v", err)
				}
			case "kill":
				container, ok := c.containerIDToContainer[event.Actor.ID]
				if !ok {
					log.Printf("missing reference for container %q. Missed event %q", event.Actor.ID, event.Action)
					continue
				}

				for k := range container.listeningPaths {
					c.watcher.Remove(k)
				}
				log.Printf("Container %q (Pod: %s) results", container.Name, container.Pod)

				if len(container.modifiedPaths) == 0 {
					log.Printf("\t Container %q (Pod %s) does not have any modified files. You can make the FS readonly", container.Name, container.Pod)
				} else {
					tree := NewTree()
					possibleRootPaths := tree.GetRootPaths(container.modifiedPaths)
					if len(possibleRootPaths) == 0 {
						log.Printf("\tThere are not possible paths because files are written at '/'")
					} else {
						for _, root := range possibleRootPaths {
							log.Printf("\tPossible to add volumes or emptydirs at %q", root)
						}
					}
				}
			}
		case err := <-errorChan:
			log.Printf("Error: %v", err)
		}
	}
}

func getPathsFromFullMergedPath(fullpath string) (mergedPath string, relativePath string) {
	mergedPath = fullpath[:pathLength]
	relativePath = fullpath[pathLength:]
	return
}

func (c *ContainerManager) WatchFiles() {
	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				mergedPath, _ := getPathsFromFullMergedPath(event.Name)
				container, ok := c.mergedPathToContainer[mergedPath]
				if !ok {
					log.Printf("Couldn't find container for merged path %q", mergedPath)
					continue
				}
				c.AddToWatcher(container, getAllSubDirectories(event.Name)...)
			case event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Chmod == fsnotify.Chmod:
				mergedPath, relativePath := getPathsFromFullMergedPath(event.Name)
				container, ok := c.mergedPathToContainer[mergedPath]
				if !ok {
					log.Printf("Couldn't find container for merged path %q", mergedPath)
					continue
				}
				if _, ok := container.modifiedPaths[relativePath]; !ok {
					log.Printf("New modification to %s in container %q (Pod %s)", relativePath, container.Name, container.Pod)
				}
				container.modifiedPaths[relativePath] = struct{}{}
			}
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (c *ContainerManager) Wait() {
	var stopChan chan bool
	<-stopChan
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	dockerClient, err := client.NewClient("unix:///host/var/run/docker.sock", api.DefaultVersion, nil, nil)
	if err != nil {
		panic(err)
	}

	cm := ContainerManager{
		watcher: watcher,
		client: dockerClient,

		containerIDToContainer: make(map[string]*Container),
		mergedPathToContainer: make(map[string]*Container),
	}

	if err := cm.Start(); err != nil {
		panic(err)
	}

	go cm.WatchDockerEvents()
	go cm.WatchFiles()
	cm.Wait()
}
