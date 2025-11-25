// Package ratelimit provides token bucket rate limiting for preventing abuse.
// This package helps protect against brute force attacks, spam, and DoS attempts
// by limiting the rate of requests per client identifier.
package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket algorithm for rate limiting.
// It tracks requests per client (identified by a string key) and enforces
// a maximum number of requests within a time window.
type RateLimiter struct {
	requests map[string][]time.Time // Map of client ID to request timestamps
	mutex    sync.Mutex             // Protects concurrent access
	limit    int                    // Maximum requests allowed
	window   time.Duration          // Time window for rate limit
}

// New creates a new RateLimiter with the specified limit and time window.
// For example: New(5, 1*time.Minute) allows 5 requests per minute.
func New(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given key should be allowed.
// Returns true if the request is within the rate limit, false otherwise.
// Automatically cleans up old request timestamps outside the window.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	requests := rl.requests[key]

	// Filter out requests outside the time window
	var recent []time.Time
	for _, t := range requests {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	// Check if limit exceeded
	if len(recent) >= rl.limit {
		// Store cleaned list but don't add new request
		rl.requests[key] = recent
		return false
	}

	// Add current request timestamp
	rl.requests[key] = append(recent, now)
	return true
}

// Reset clears all rate limit data for a specific key.
// Useful for resetting limits after successful authentication or as admin action.
func (rl *RateLimiter) Reset(key string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	delete(rl.requests, key)
}

// ResetAll clears all rate limit data.
// Useful for periodic cleanup or maintenance.
func (rl *RateLimiter) ResetAll() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.requests = make(map[string][]time.Time)
}

// GetCount returns the current request count for a key within the window.
// Useful for monitoring and debugging.
func (rl *RateLimiter) GetCount(key string) int {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	requests := rl.requests[key]
	count := 0

	for _, t := range requests {
		if t.After(cutoff) {
			count++
		}
	}

	return count
}

// CleanupOldEntries removes old request data to prevent memory growth.
// Should be called periodically (e.g., every hour) for long-running servers.
func (rl *RateLimiter) CleanupOldEntries() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for key, requests := range rl.requests {
		var recent []time.Time
		for _, t := range requests {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = recent
		}
	}
}
