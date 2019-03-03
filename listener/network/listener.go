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

	networkCalls map[networkKey]struct{}

	lock sync.Mutex
}

func NewListener() *Listener {
	return &Listener{
		output: make(chan types.Network),

		networkCalls: make(map[networkKey]struct{}),
	}
}

func (l *Listener) Output() <-chan types.Network {
	return l.output
}

type networkKey struct {
	Command, SAddr, DAddr string
	DPort                 int
}

func networkKeyFromNetwork(n types.Network) networkKey {
	return networkKey{
		Command: n.Command,
		SAddr:   n.SAddr,
		DAddr:   n.DAddr,
		DPort:   n.DPort,
	}
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

	nc := types.Network{
		PID:     pid,
		Command: values[1],
		SAddr:   values[3],
		DAddr:   values[4],
		DPort:   port,
		Call:    call,
	}

	mapKey := networkKeyFromNetwork(nc)
	if _, ok := l.networkCalls[mapKey]; ok {
		return
	}
	l.networkCalls[mapKey] = struct{}{}

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
	go l.start("tcpconnect", "accept")
}
