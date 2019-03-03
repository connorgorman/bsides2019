package main

import (
	"log"

	"github.com/connorgorman/bsides2019/listener/capable"
	"github.com/connorgorman/bsides2019/listener/docker"
	"github.com/connorgorman/bsides2019/listener/fsmonitor"
	"github.com/connorgorman/bsides2019/listener/pid"
)

func main() {
	dockerListener, err := docker.NewListener()
	if err != nil {
		panic(err)
	}
	go dockerListener.Start()

	capableListener := capable.NewListener()
	go capableListener.Start()

	pidListener := pid.NewListener()
	go pidListener.Start()

	fileListener, err := fsmonitor.NewListener()
	if err != nil {
		panic(err)
	}
	go fileListener.Start()

	client := newClient("http://server.bsides:8080")
	for {
		select {
		case container := <-dockerListener.NewContainerChannel():
			fileListener.AddContainer(&container)
			client.SendContainer(container)

		case cid := <-dockerListener.RemoveContainerChannel():
			fileListener.RemoveContainer(cid)

		case cap := <-capableListener.Output():
			log.Printf("Capability: %+v", cap)
			client.SendCapability(cap)

		case pid := <-pidListener.Output():
			capableListener.AddContainer(pid)

		case file := <-fileListener.Output():
			client.SendFile(file)
		}
	}
}
