package fsmonitor

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/connorgorman/bsides2019/types"
	"github.com/fsnotify/fsnotify"
)

var pathsToIgnore = map[string]struct{}{
	"proc":  {},
	"dev":   {},
	"usr":   {},
	"boot":  {},
	"lib":   {},
	"lib64": {},
	"sys":   {},
}

var pathLength = 0

type containerWrapper struct {
	*types.Container

	modifiedPaths  map[string]struct{}
	listeningPaths map[string]struct{}
}

func newContainerWrapper(c *types.Container) *containerWrapper {
	return &containerWrapper{
		Container: c,

		modifiedPaths:  make(map[string]struct{}),
		listeningPaths: make(map[string]struct{}),
	}
}

type Listener struct {
	watcher *fsnotify.Watcher

	containerIDToContainer map[string]*containerWrapper
	mergedPathToContainer  map[string]*containerWrapper

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
		mergedPathToContainer:  make(map[string]*containerWrapper),

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

	dirs, err := ioutil.ReadDir(wrap.FilePath)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fullpaths := []string{wrap.FilePath}
	for _, d := range dirs {
		if _, ok := pathsToIgnore[d.Name()]; ok {
			continue
		}
		fullpaths = append(fullpaths, getSubFiles(filepath.Join(wrap.FilePath, d.Name()))...)
	}

	c.AddToWatcher(wrap, fullpaths...)
}

func (c *Listener) AddToWatcher(container *containerWrapper, files ...string) {
	for _, f := range files {
		if err := c.watcher.Add(f); err != nil {
			log.Printf("error adding file %q to watcher: %v", f, err)
			continue
		}
		container.listeningPaths[f] = struct{}{}
	}
}

func getSubFiles(path string) []string {
	fi, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if !fi.IsDir() {
		return nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	dirs := []string{path}
	for _, f := range files {
		dirs = append(dirs, getSubFiles(filepath.Join(path, f.Name()))...)
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
				log.Printf("events has returned")
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
				c.AddToWatcher(container, getSubFiles(event.Name)...)
				fallthrough
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
					Path:        relativePath,
				}
			}
		case err, ok := <-c.watcher.Errors:
			if !ok {
				log.Printf("errors has returned")
				return
			}
			log.Println("error:", err)
		}
	}
}
