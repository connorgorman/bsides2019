package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

type Server struct {

}



func (s *Server) Start() {
	r := mux.NewRouter()

	r.HandleFunc("/files", s.FilesPostHandler)
	r.HandleFunc("/files/{id}", s.FilesHandler)

	r.HandleFunc("/capabilities/{id}", s.CapabilitiesHandler)
	r.HandleFunc("/capabilities}", s.CapabilitiesPostHandler)
}

func (s *Server) FilesPostHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) FilesHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) CapabilitiesPostHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) CapabilitiesHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *Server) NetworkHandler(w http.ResponseWriter, req *http.Request) {

}


func main() {

	r := mux.NewRouter()
	s := r.PathPrefix("/products").Subrouter()
	// "/products/"
	s.HandleFunc("/", ProductsHandler)
	// "/products/{key}/"
	s.HandleFunc("/{key}/", ProductHandler)
	// "/products/{key}/details"
	s.HandleFunc("/{key}/details", ProductDetailsHandler)







	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to my website!")
	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":80", nil)
}

