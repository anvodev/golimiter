package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type TokenBucket struct {
	rate       float64
	burst      int
	tokens     int
	lastUpdate time.Time
}

func NewTokenBucket(r float64, b int) *TokenBucket {
	return &TokenBucket{
		rate:       r,
		burst:      b,
		tokens:     b,
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	currentTime := time.Now()
	elapsed := currentTime.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = currentTime
	tb.tokens += int(tb.rate * elapsed)
	if tb.tokens > tb.burst {
		tb.tokens = tb.burst
	}
	if tb.tokens > 0 {
		tb.tokens -= 1
		return true
	} else {
		return false
	}
}

func rateLimiter(next http.Handler) http.Handler {
	// do something globally
	tb := NewTokenBucket(0.5, 2)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do something when the request is called
		if !tb.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limited"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
