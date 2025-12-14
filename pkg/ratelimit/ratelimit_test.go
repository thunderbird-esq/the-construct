package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	rl := New(5, time.Minute)
	if rl == nil {
		t.Fatal("New returned nil")
	}
	if rl.limit != 5 {
		t.Errorf("limit = %d, want 5", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("window = %v, want 1m", rl.window)
	}
	if rl.requests == nil {
		t.Error("requests map not initialized")
	}
}

func TestAllow(t *testing.T) {
	rl := New(3, time.Second)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-client") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("test-client") {
		t.Error("4th request should be denied")
	}

	// Different client should be allowed
	if !rl.Allow("other-client") {
		t.Error("Different client should be allowed")
	}
}

func TestAllowWindowExpiry(t *testing.T) {
	rl := New(2, 50*time.Millisecond)

	// Use up the limit
	rl.Allow("client")
	rl.Allow("client")

	// Should be denied
	if rl.Allow("client") {
		t.Error("Should be denied at limit")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("client") {
		t.Error("Should be allowed after window expiry")
	}
}

func TestReset(t *testing.T) {
	rl := New(2, time.Minute)

	// Use up the limit
	rl.Allow("client")
	rl.Allow("client")

	if rl.Allow("client") {
		t.Error("Should be denied at limit")
	}

	// Reset the client
	rl.Reset("client")

	// Should be allowed again
	if !rl.Allow("client") {
		t.Error("Should be allowed after reset")
	}
}

func TestResetAll(t *testing.T) {
	rl := New(1, time.Minute)

	// Add requests for multiple clients
	rl.Allow("client1")
	rl.Allow("client2")
	rl.Allow("client3")

	// Reset all
	rl.ResetAll()

	// All clients should be allowed again
	if !rl.Allow("client1") {
		t.Error("client1 should be allowed after reset all")
	}
	if !rl.Allow("client2") {
		t.Error("client2 should be allowed after reset all")
	}
}

func TestGetCount(t *testing.T) {
	rl := New(10, time.Minute)

	if count := rl.GetCount("client"); count != 0 {
		t.Errorf("count = %d, want 0 for new client", count)
	}

	rl.Allow("client")
	rl.Allow("client")
	rl.Allow("client")

	if count := rl.GetCount("client"); count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestCleanupOldEntries(t *testing.T) {
	rl := New(10, 50*time.Millisecond)

	// Add some requests
	rl.Allow("client1")
	rl.Allow("client2")

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Cleanup
	rl.CleanupOldEntries()

	// Counts should be 0
	if count := rl.GetCount("client1"); count != 0 {
		t.Errorf("client1 count = %d after cleanup, want 0", count)
	}
}

func TestConcurrency(t *testing.T) {
	rl := New(100, time.Second)
	var wg sync.WaitGroup

	// Spawn multiple goroutines making requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rl.Allow("shared-client")
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and should have limited requests
	count := rl.GetCount("shared-client")
	if count > 100 {
		t.Errorf("count = %d, should not exceed limit of 100", count)
	}
}
