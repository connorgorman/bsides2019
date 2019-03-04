package network

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/connorgorman/bsides2019/types"
)

type Listener struct {
	output chan types.Network

	networkCalls map[types.Network]struct{}

	lock sync.Mutex
}

func NewListener() *Listener {
	return &Listener{
		output: make(chan types.Network),

		networkCalls: make(map[types.Network]struct{}),
	}
}

func (l *Listener) Output() <-chan types.Network {
	return l.output
}

func (l *Listener) parseAndOutput(line string, call string) {
	values := strings.Fields(line)
	if len(values) < 6 {
		return
	}

	pid, err := strconv.Atoi(values[0])
	if err != nil {
		log.Printf("could not parse pid: %q", values[0])
		return
	}

	port, err := strconv.Atoi(values[5])
	if err != nil {
		log.Printf("could not parse port: %q", values[5])
		return
	}

	src := strings.TrimPrefix(values[3], "::ffff:")
	dst := strings.TrimPrefix(values[4], "::ffff:")
	if src == "127.0.0.1" || dst == "127.0.0.1" {
		return
	}

	nc := types.Network{
		PID:     pid,
		Command: values[1],
		SAddr:   src,
		DAddr:   dst,
		DPort:   port,
		Call:    call,
	}

	l.lock.Lock()
	if _, ok := l.networkCalls[nc]; ok {
		l.lock.Unlock()
		return
	}
	l.networkCalls[nc] = struct{}{}
	l.lock.Unlock()

	l.output <- nc
}

func (l *Listener) start(process, call string) {
	cmd := exec.Command(fmt.Sprintf("/usr/share/bcc/tools/%s", process))
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

		l.parseAndOutput(string(line), call)
	}
}

func (l *Listener) Start() {
	go l.start("tcpconnect", "connect")
	go l.start("tcpaccept", "accept")
}
