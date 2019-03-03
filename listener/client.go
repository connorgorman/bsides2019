package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/connorgorman/bsides2019/types"
)

type client struct {
	*http.Client

	endpoint string
}

func newClient(endpoint string) *client {
	return &client{
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
		endpoint: endpoint,
	}
}

func (c *client) SendContainer(container types.Container) {
	if err := c.sendRequest("/containers", &container); err != nil {
		log.Printf("error sending containers: %v", err)
	}
}

func (c *client) SendFile(file types.File) {
	if err := c.sendRequest("/files", &file); err != nil {
		log.Printf("error sending files: %v", err)
	}
}

func (c *client) SendCapability(capability types.Capability) {
	if err := c.sendRequest("/capabilities", &capability); err != nil {
		log.Printf("error sending capability: %v", err)
	}
}

func (c *client) sendRequest(url string, i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.endpoint+url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = c.Do(req)
	return err
}
