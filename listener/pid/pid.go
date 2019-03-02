package pid

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type ContainerPID struct {
	ID  string
	PID int
	Command string
}

type Listener struct {
	pidChan chan ContainerPID
}

func NewListener() *Listener {
	return &Listener{
		pidChan: make(chan ContainerPID),
	}
}

func (l *Listener) Output() <-chan ContainerPID {
	return l.pidChan
}

func (l *Listener) parseCgroupAndOutput(pid int) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/host/proc/%d/cgroup", pid))
	if err != nil {
		return
	}

	command, err := ioutil.ReadFile(fmt.Sprintf("/host/proc/%d/cmdline", pid))
	if err != nil {
		return
	}

	dataStr := string(data)
	line := strings.SplitN(dataStr, "\n", 2)[0]
	lineSplit := strings.Split(line, "/")
	containerID := lineSplit[len(lineSplit)-1]
	if len(containerID) != 64 {
		return
	}
	l.pidChan <- ContainerPID{ID: containerID, PID: pid, Command: string(command)}
}

func (l *Listener) Start() {
	path := "/host/proc"

	t := time.NewTicker(10 * time.Millisecond)

	seenPids := make(map[int]struct{})
	for {
		<-t.C
		dirs, err := ioutil.ReadDir(path)
		if err != nil {
			log.Printf("error getting dirs: %v", err)
			continue
		}
		for _, d := range dirs {
			pid, err := strconv.Atoi(d.Name())
			if err != nil {
				continue
			}
			if _, ok := seenPids[pid]; ok {
				continue
			}
			seenPids[pid] = struct{}{}
			l.parseCgroupAndOutput(pid)
		}
	}
}