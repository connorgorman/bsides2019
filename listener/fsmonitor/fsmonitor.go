package fsmonitor

import (
	"github.com/connorgorman/bsides2019/types"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

var pathsToIgnore = map[string]struct{}{
	"proc":{},
	"dev":{},
}

var pathLength = 0

type containerWrapper struct {
	*types.Container

	modifiedPaths map[string]struct{}
	listeningPaths map[string]struct{}
}

func newContainerWrapper(c *types.Container) *containerWrapper {
	return &containerWrapper{
		Container: c,

		modifiedPaths: make(map[string]struct{}),
		listeningPaths: make(map[string]struct{}),
	}
}

type Listener struct {
	watcher *fsnotify.Watcher

	containerIDToContainer map[string]*containerWrapper
	mergedPathToContainer map[string]*containerWrapper

	output chan types.File
}

 func NewListener() (*Listener, error) {
	 watcher, err := fsnotify.NewWatcher()
	 if err != nil {
	 	return nil, err
	 }
	return &Listener{
		watcher: watcher,

		containerIDToContainer: make(map[string]*containerWrapper),
		mergedPathToContainer: make(map[string]*containerWrapper),

		output: make(chan types.File),
	}, nil
 }

func (c *Listener) Start() {
	c.watchFiles()
}

func (c *Listener) Output() <-chan types.File {
	return c.output
}

func (c *Listener) AddContainer(container *types.Container) {
	wrap := newContainerWrapper(container)
	c.containerIDToContainer[container.ID] = wrap

	// Set the index of the prefix for the container filesystems
	// this allows for quick indexing when getting the filepaths within the container
	if pathLength == 0 {
		pathLength = len(wrap.FilePath)
	}
	c.mergedPathToContainer[wrap.FilePath] = wrap
	dirs := getAllSubDirectories(wrap.FilePath)
	c.AddToWatcher(wrap, dirs...)
}

func (c *Listener) AddToWatcher(container *containerWrapper, files ...string) {
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

func (c *Listener) RemoveContainer(id string) {
	container, ok := c.containerIDToContainer[id]
	if !ok {
		log.Printf("missing reference for container %q", id)
		return
	}

	for k := range container.listeningPaths {
		c.watcher.Remove(k)
	}
}

func getPathsFromFullMergedPath(fullpath string) (mergedPath string, relativePath string) {
	mergedPath = fullpath[:pathLength]
	relativePath = fullpath[pathLength:]
	return
}

func (c *Listener) watchFiles() {
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
				if _, ok := container.modifiedPaths[relativePath]; ok {
					continue
				}
				container.modifiedPaths[relativePath] = struct{}{}
				c.output <- types.File{
					ContainerID: container.ID,
					Path: relativePath,
				}
			}
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
