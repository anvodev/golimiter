package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type TokenBucket struct {
	rate       float64   // How many tokens are refilled every second
	burst      int       // Maximum capacity of the bucket
	tokens     int       // Current tokens in the bucket
	lastUpdate time.Time // Last time we update the tokens amount
}

func NewTokenBucket(r float64, b int) *TokenBucket {
	return &TokenBucket{
		rate:       r,
		burst:      b,
		tokens:     b, // Initialize the tokens at max capacity of the bucket
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	currentTime := time.Now()
	elapsed := currentTime.Sub(tb.lastUpdate).Seconds() // Calculate the time since the last refill
	tb.tokens += int(tb.rate * elapsed)                 // Refill the tokens base on the rate and the time from the last refill
	tb.lastUpdate = currentTime                         // When the next request comes, we will calculate the refill tokens based on this update
	if tb.tokens > tb.burst {                           // Burst is the maximum capacity of the bucket
		tb.tokens = tb.burst
	}
	if tb.tokens > 0 { // There is some available tokens, so the request is passed and 1 token is consumed
		tb.tokens -= 1
		return true
	} else { // There is no token left, and the request is dropped.
		return false
	}
}

func rateLimiter(next http.Handler) http.Handler {
	// Do something globally, initialization
	tb := NewTokenBucket(0.5, 2) // New tokens are replenished at a rate of 0.5 tokens per second. The maximum burst request amount is 2

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do something for each request
		if !tb.Allow() {
			w.Header().Set("Content-Type", "application/json")
			// We can also use some additional headers such as X-Ratelimit-Remaining, X-Ratelimit-Limit, and X-Ratelimit-Retry-After for further clarification
			w.WriteHeader(http.StatusTooManyRequests) // 429 StatusCode
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limited"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
