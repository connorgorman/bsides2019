package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Server struct {

}

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/containers", s.ContainerPostHandler)
	r.HandleFunc("/containers/{id}", s.ContainerHandler)

	r.HandleFunc("/files", s.FilesPostHandler)
	r.HandleFunc("/files/{id}", s.FilesHandler)

	r.HandleFunc("/capabilities/{id}", s.CapabilitiesHandler)
	r.HandleFunc("/capabilities}", s.CapabilitiesPostHandler)
	return r
}

func (s *Server) ContainerPostHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Container")
}

func (s *Server) ContainerHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) FilesPostHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Files")

}

func (s *Server) FilesHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) CapabilitiesPostHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Capabilities")

}

func (s *Server) CapabilitiesHandler(w http.ResponseWriter, req *http.Request) {

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

