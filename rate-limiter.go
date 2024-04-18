package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type TokenBucket struct {
	rate       float64   // how many tokens are refilled every second
	burst      int       // maximum capacity of the bucket
	tokens     int       // current tokens in the bucket
	lastUpdate time.Time // use last request time as the start time of current period
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
	tb.tokens += int(tb.rate * elapsed) // refill the tokens base on the rate and the time from last request
	if tb.tokens > tb.burst {           // burst is the maximum capacity of the bucket
		tb.tokens = tb.burst
	}
	if tb.tokens > 0 { // there is available token, the request is passed
		tb.tokens -= 1
		return true
	} else { // there is no token left, the request is dropped
		return false
	}
}

func rateLimiter(next http.Handler) http.Handler {
	// do something globally, initialization
	tb := NewTokenBucket(0.5, 2) // new tokens are filled at the rate 0.5 token per second. Max burst request amount is 2

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do something when the request is called
		if !tb.Allow() {
			w.Header().Set("Content-Type", "application/json")
			// We can also use some additional headers like X-Ratelimit-Remaining, X-Ratelimit-Limit, X-Ratelimit-Retry-After headers for clarification
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limited"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
