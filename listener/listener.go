package main

import (
	"github.com/connorgorman/bsides2019/listener/capable"
	"github.com/connorgorman/bsides2019/listener/docker"
	"github.com/connorgorman/bsides2019/listener/pid"
	"log"
)

func main(){
	dockerListener, err := docker.NewListener()
	if err != nil {
		panic(err)
	}
	go dockerListener.Start()

	capableListener := capable.NewListener()
	//go capableListener.Start()

	pidListener := pid.NewListener()
	go pidListener.Start()

	for {
		select {
		case container := <-dockerListener.NewContainerChannel():
			log.Printf("New container add: %+v", container)
			//capableListener.AddContainer(container)
		case cid := <-dockerListener.RemoveContainerChannel():
			log.Printf("Removed %q", cid)
		case cap := <-capableListener.Output():
			log.Printf("Cap: %+v", cap)
		case pid := <-pidListener.Output():
			capableListener.AddContainer(pid)
		}
	}

}
