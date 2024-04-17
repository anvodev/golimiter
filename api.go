package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string
}

func WriteJSON(w http.ResponseWriter, v any) {
	json.NewEncoder(w).Encode(v)
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			WriteJSON(w, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	server := mux.NewRouter()
	server.HandleFunc("/ping", makeHTTPHandlerFunc(handlePing))

	fmt.Println("server http on port ", s.listenAddr)
	panic(http.ListenAndServe(s.listenAddr, server))
}

func handlePing(w http.ResponseWriter, r *http.Request) error {
	return fmt.Errorf("error sml")
}
