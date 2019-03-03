package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
}

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/containers", s.ContainerPostHandler)
	r.HandleFunc("/files", s.FilesPostHandler)
	r.HandleFunc("/capabilities", s.CapabilitiesPostHandler)
	r.HandleFunc("/pids", s.PIDsPostHandler)

	return r
}

func (s *Server) ContainerPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) FilesPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) CapabilitiesPostHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	log.Println(string(data))
}

func (s *Server) PIDsPostHandler(w http.ResponseWriter, req *http.Request) {
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
