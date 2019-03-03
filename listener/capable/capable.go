package capable

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/connorgorman/bsides2019/types"
)

type Listener struct {
	output    chan types.Capability
	pidToCaps map[int]map[string]struct{}

	lock sync.Mutex
}

func NewListener() *Listener {
	return &Listener{
		output:    make(chan types.Capability),
		pidToCaps: make(map[int]map[string]struct{}),
	}
}

func (l *Listener) Output() <-chan types.Capability {
	return l.output
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

	if _, ok := l.pidToCaps[pid]; !ok {
		l.pidToCaps[pid] = make(map[string]struct{})
	}
	cap := values[6]
	if _, ok := l.pidToCaps[pid][cap]; ok {
		return
	}
	l.pidToCaps[pid][cap] = struct{}{}
	l.output <- types.Capability{
		PID:     pid,
		Command: values[4],
		Cap:     cap,
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
