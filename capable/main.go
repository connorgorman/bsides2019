package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Line struct {
	pid          int
	cap, command string
}

func parseLine(line string) (*Line, error) {
	values := strings.Fields(line)
	if len(values) < 8 {
		log.Printf("Invalid number of values")
		return nil, fmt.Errorf("invalid number of values: %d", len(values))
	}
	// if audit == 0, then ignore
	if values[7] == "0" {
		return nil, nil
	}
	pid, err := strconv.Atoi(values[2])
	if err != nil {
		return nil, err
	}
	return &Line{
		pid:     pid,
		command: values[4],
		cap:     values[6],
	}, nil
}

func main() {
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

		parsedLine, err := parseLine(string(line))
		if err != nil {
			log.Printf("Error parsing line %q: %v", line, err)
			continue
		}
		log.Printf("Capable line: %+v", parsedLine)
	}

}
