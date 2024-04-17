package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusInternalServerError, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
	counter    int64
	rateLimit  int64
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		counter:    4,
		rateLimit:  4,
	}
}

func (s *APIServer) handleRateLimit(next http.Handler) http.Handler {
	ticker := time.NewTicker(5 * time.Second) // fill 2 tokens every 5 seconds
	go func() {
		for {
			t := <-ticker.C
			fmt.Println("Fill up the Rate Limiter Counter", t)
			filled := s.counter + 2
			s.counter = int64(min(float64(s.rateLimit), float64(filled)))
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-limit", strconv.FormatInt(s.rateLimit, 10))
		if s.counter == 0 {
			w.Header().Set("X-Ratelimit-Retry-After", "forever")
			w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(int(s.counter)))
			WriteJSON(w, http.StatusTooManyRequests, ApiError{
				Error: "Rate limited",
			})
		} else {
			s.counter--
			w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(int(s.counter)))
			next.ServeHTTP(w, r)
		}
	})
}

func (s *APIServer) Run() {
	server := mux.NewRouter()
	server.HandleFunc("/ping", makeHTTPHandlerFunc(handlePing))

	fmt.Println("server http on port ", s.listenAddr)

	panic(http.ListenAndServe(s.listenAddr, s.handleRateLimit(server)))
}

func handlePing(w http.ResponseWriter, r *http.Request) error {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "OK"})
	return nil
}
