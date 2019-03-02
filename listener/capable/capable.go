package capable

import (
	"bufio"
	"github.com/connorgorman/bsides2019/listener/pid"
	"github.com/connorgorman/bsides2019/types"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Listener struct {
	output chan *types.Capability
	pidsToContainers map[int]string
	containerToCaps map[string]map[string]struct{}

	lock sync.Mutex
}

func NewListener() *Listener {
	return &Listener{
		output: make(chan *types.Capability),
		pidsToContainers: make(map[int]string),
		containerToCaps: make(map[string]map[string]struct{}),
	}
}

func (l *Listener) Output() <-chan *types.Capability {
	return l.output
}

func (l *Listener) AddContainer(c pid.ContainerPID) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.pidsToContainers[c.PID] = c.ID
}

func (l *Listener) parseAndOutput(line string) {
	values := strings.Fields(line)
	if len(values) < 8 {
		return
	}
	// if audit == 0, then ignore
	if values[7] == "0" {
		return
	}
	pid, err := strconv.Atoi(values[2])
	if err != nil {
		log.Printf("could not parse pid: %q", values[2])
		return
	}

	// Delay so that pid can be scraped
	time.Sleep(10 * time.Millisecond)
	l.lock.Lock()
	cid, ok := l.pidsToContainers[pid]
	l.lock.Unlock()
	if !ok {
		log.Printf("dropping capability: %d - %s", pid, values[4])
		return
	}

	if _, ok := l.containerToCaps[cid]; !ok {
		l.containerToCaps[cid] = make(map[string]struct{})
	}
	l.output <- &types.Capability{
		ContainerID: cid,
		PID: pid,
		Command: values[4],
		Cap: values[6],
	}
}

func (l *Listener) Start() {
	cmd := exec.Command("/usr/share/bcc/tools/capable")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(stdoutPipe)
	// ignore header
	_, _, err = reader.ReadLine()
	if err != nil {
		panic(err)
	}
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			log.Printf("EOF: %v", err)
			return
		} else if err != nil {
			panic(err)
		}

		l.parseAndOutput(string(line))
	}
}
