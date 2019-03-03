package main

import (
	"encoding/json"
	"github.com/connorgorman/bsides2019/types"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	containerMap map[string]*types.Container
	containerToPidMap map[string][]*types.ContainerPID
	containerToFilesMap map[string][]string
	pidsToCaps          map[int][]*types.Capability

	lock sync.RWMutex
}

type FileResponse struct {
	PotentialFSRoots []string
	ReadOnlyPossible bool
}

type ContainerResponse struct {
	Container *types.Container
	CapabilitiesRequired []*types.Capability
	File FileResponse
}

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/containers", s.ContainerPostHandler).Methods("POST")
	r.HandleFunc("/files", s.FilesPostHandler)
	r.HandleFunc("/capabilities", s.CapabilitiesPostHandler)
	r.HandleFunc("/pids", s.PIDsPostHandler)
	r.HandleFunc("/containers", s.ContainerGetHandler).Methods("GET")
	return r
}

func (s *Server) ContainerGetHandler(w http.ResponseWriter, req *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	containerResponses := make([]ContainerResponse, 0, len(s.containerMap))

	for cid, container := range s.containerMap {
		pids := s.containerToPidMap[cid]
		var capabilities []*types.Capability
		for _, p := range pids {
			capabilities = append(capabilities, s.pidsToCaps[p.PID]...)
		}

		roots, possible := s.containerToFilesMap[cid]
		containerResponses = append(containerResponses, ContainerResponse{
			Container: container,
			CapabilitiesRequired: capabilities,
			File: FileResponse{
				PotentialFSRoots: roots,
				ReadOnlyPossible: possible,
			},
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

func (s *Server) ContainerPostHandler(w http.ResponseWriter, req *http.Request) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) FilesPostHandler(w http.ResponseWriter, req *http.Request) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) CapabilitiesPostHandler(w http.ResponseWriter, req *http.Request) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) PIDsPostHandler(w http.ResponseWriter, req *http.Request) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) NetworkHandler(w http.ResponseWriter, req *http.Request) {

}

func main() {
	var server Server
	srv := &http.Server{
		Handler:      server.GetRouter(),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
