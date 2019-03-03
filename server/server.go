package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/connorgorman/bsides2019/types"
	"github.com/gorilla/mux"
)

type server struct {
	containerMap        map[string]*types.Container
	containerToPidMap   map[string][]*types.ContainerPID
	containerToFilesMap map[string][]string
	pidsToCaps          map[int][]*types.Capability
	pidsToNetwork       map[int][]*types.Network

	lock sync.Mutex
}

func newServer() *server {
	return &server{
		containerMap:        make(map[string]*types.Container),
		containerToPidMap:   make(map[string][]*types.ContainerPID),
		containerToFilesMap: make(map[string][]string),

		pidsToCaps:    make(map[int][]*types.Capability),
		pidsToNetwork: make(map[int][]*types.Network),
	}
}

type FileResponse struct {
	PotentialFSRoots []string `json:",omitempty"`
	ReadOnlyPossible bool
}

type ContainerResponse struct {
	Container            *types.Container    `json:",omitempty"`
	CapabilitiesRequired []*types.Capability `json:",omitempty"`
	File                 FileResponse        `json:",omitempty"`
	Network              []*types.Network    `json:",omitempty"`
}

func (s *server) GetRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/containers", s.ContainerPostHandler).Methods("POST")
	r.HandleFunc("/containers", s.ContainerGetHandler).Methods("GET")
	r.HandleFunc("/containers/{name}", s.ContainerGetHandler).Methods("GET")

	r.HandleFunc("/files", s.FilesPostHandler)
	r.HandleFunc("/capabilities", s.CapabilitiesPostHandler)
	r.HandleFunc("/pids", s.PIDsPostHandler)
	r.HandleFunc("/network", s.NetworkPostHandler)
	return r
}

type capabilityKey struct {
	cap, command string
}

func (s *server) ContainerGetHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namePrefix := vars["name"]

	s.lock.Lock()
	defer s.lock.Unlock()

	containerResponses := make([]ContainerResponse, 0, len(s.containerMap))
	for cid, container := range s.containerMap {
		if !strings.HasPrefix(container.Name, namePrefix) {
			continue
		}
		// For the sake of the demo, ignore kube-system
		if container.Namespace == "kube-system" {
			continue
		}
		pids := s.containerToPidMap[cid]
		var capabilities []*types.Capability
		var network []*types.Network

		seenCap := make(map[capabilityKey]struct{})
		for _, p := range pids {
			for _, c := range s.pidsToCaps[p.PID] {
				if _, ok := seenCap[capabilityKey{command: c.Command, cap: c.Cap}]; ok {
					continue
				}
				seenCap[capabilityKey{command: c.Command, cap: c.Cap}] = struct{}{}
				capabilities = append(capabilities, c)
			}

			log.Printf("%s -> %d", container.Name, p.PID)
			network = append(network, s.pidsToNetwork[p.PID]...)
		}

		roots, possible := GetRootPaths(s.containerToFilesMap[cid])
		containerResponses = append(containerResponses, ContainerResponse{
			Container:            container,
			CapabilitiesRequired: capabilities,
			File: FileResponse{
				PotentialFSRoots: roots,
				ReadOnlyPossible: possible,
			},
			Network: network,
		})
	}

	bytes, err := json.Marshal(containerResponses)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	if _, err := w.Write(bytes); err != nil {
		log.Printf("error writing: %v", err)
	}
}

func (s *server) ContainerPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	var container types.Container
	if err := json.Unmarshal(data, &container); err != nil {
		log.Printf("error unmarshalling container: %v", err)
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.containerMap[container.ID] = &container
}

func (s *server) FilesPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	var file types.File
	if err := json.Unmarshal(data, &file); err != nil {
		log.Printf("error unmarshalling file: %v", err)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.containerToFilesMap[file.ContainerID] = append(s.containerToFilesMap[file.ContainerID], file.Path)
}

func (s *server) CapabilitiesPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	var capability types.Capability
	if err := json.Unmarshal(data, &capability); err != nil {
		log.Printf("error unmarshalling capability: %v", err)
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.pidsToCaps[capability.PID] = append(s.pidsToCaps[capability.PID], &capability)
}

func (s *server) PIDsPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}
	var pid types.ContainerPID
	if err := json.Unmarshal(data, &pid); err != nil {
		log.Printf("error unmarshalling pid: %v", err)
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.containerToPidMap[pid.ID] = append(s.containerToPidMap[pid.ID], &pid)
}

func (s *server) NetworkPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}
	var network types.Network
	if err := json.Unmarshal(data, &network); err != nil {
		log.Printf("error unmarshalling network: %v", err)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pidsToNetwork[network.PID] = append(s.pidsToNetwork[network.PID], &network)
}

func main() {
	server := newServer()
	srv := &http.Server{
		Handler:      server.GetRouter(),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
