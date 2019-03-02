package main

import (
	"github.com/stackrox/demo/listener/capable"
	"github.com/stackrox/demo/listener/docker"
	"log"
)

func main(){
	dockerListener, err := docker.NewListener()
	if err != nil {
		panic(err)
	}
	go dockerListener.Start()

	capableListener := capable.NewListener()
	go capableListener.Start()

	for {
		select {
		case container := <-dockerListener.NewContainerChannel():
			capableListener.AddContainer(container)
		case cid := <-dockerListener.RemoveContainerChannel():
			log.Printf("Removed %q", cid)
		case cap := <-capableListener.Output():
			log.Printf("Cap: %+v", cap)
		}
	}

}
