package events

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWebhookManager(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	if wm == nil {
		t.Fatal("NewWebhookManager returned nil")
	}
}

func TestWebhookAddRemove(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	config := &WebhookConfig{
		ID:      "test-webhook",
		Name:    "Test Webhook",
		URL:     "http://example.com/webhook",
		Enabled: true,
	}

	err := wm.AddWebhook(config)
	if err != nil {
		t.Fatalf("AddWebhook failed: %v", err)
	}

	wh := wm.GetWebhook("test-webhook")
	if wh == nil {
		t.Error("GetWebhook should return webhook")
	}
	if wh.Name != "Test Webhook" {
		t.Errorf("Name = %s, want Test Webhook", wh.Name)
	}

	wm.RemoveWebhook("test-webhook")
	wh = wm.GetWebhook("test-webhook")
	if wh != nil {
		t.Error("Webhook should be removed")
	}
}

func TestWebhookValidation(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	// Missing ID
	err := wm.AddWebhook(&WebhookConfig{URL: "http://example.com"})
	if err == nil {
		t.Error("Should fail with missing ID")
	}

	// Missing URL
	err = wm.AddWebhook(&WebhookConfig{ID: "test"})
	if err == nil {
		t.Error("Should fail with missing URL")
	}
}

func TestWebhookDefaults(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	config := &WebhookConfig{
		ID:  "test",
		URL: "http://example.com",
	}
	wm.AddWebhook(config)

	if config.RetryCount != 3 {
		t.Errorf("RetryCount = %d, want 3", config.RetryCount)
	}
	if config.Timeout != 10 {
		t.Errorf("Timeout = %d, want 10", config.Timeout)
	}
}

func TestWebhookEnableDisable(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	wm.AddWebhook(&WebhookConfig{
		ID:      "test",
		URL:     "http://example.com",
		Enabled: true,
	})

	err := wm.DisableWebhook("test")
	if err != nil {
		t.Fatalf("DisableWebhook failed: %v", err)
	}

	wh := wm.GetWebhook("test")
	if wh.Enabled {
		t.Error("Webhook should be disabled")
	}

	err = wm.EnableWebhook("test")
	if err != nil {
		t.Fatalf("EnableWebhook failed: %v", err)
	}

	wh = wm.GetWebhook("test")
	if !wh.Enabled {
		t.Error("Webhook should be enabled")
	}
}

func TestWebhookEnableNotFound(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	err := wm.EnableWebhook("nonexistent")
	if err == nil {
		t.Error("Should fail for non-existent webhook")
	}
}

func TestWebhookListWebhooks(t *testing.T) {
	eb := NewEventBus(2)
	wm := NewWebhookManager(eb, "test-server")

	wm.AddWebhook(&WebhookConfig{ID: "wh1", URL: "http://example1.com"})
	wm.AddWebhook(&WebhookConfig{ID: "wh2", URL: "http://example2.com"})
	wm.AddWebhook(&WebhookConfig{ID: "wh3", URL: "http://example3.com"})

	webhooks := wm.ListWebhooks()
	if len(webhooks) != 3 {
		t.Errorf("Expected 3 webhooks, got %d", len(webhooks))
	}
}

func TestWebhookDelivery(t *testing.T) {
	// Create test server
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type should be application/json")
		}
		if r.Header.Get("X-Event-Type") == "" {
			t.Error("X-Event-Type header should be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{
		ID:      "test",
		URL:     server.URL,
		Enabled: true,
		Events:  []EventType{EventPlayerJoin},
	})

	eb.Publish(NewEvent(EventPlayerJoin).WithPlayer("Test", 1))

	// Wait for delivery
	time.Sleep(100 * time.Millisecond)

	if !received {
		t.Error("Webhook should have been delivered")
	}
}

func TestWebhookEventFilter(t *testing.T) {
	var received int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{
		ID:      "test",
		URL:     server.URL,
		Enabled: true,
		Events:  []EventType{EventPlayerJoin}, // Only join events
	})

	eb.Publish(NewEvent(EventPlayerJoin))
	eb.Publish(NewEvent(EventPlayerLeave))  // Should be filtered
	eb.Publish(NewEvent(EventNPCKill))      // Should be filtered

	time.Sleep(100 * time.Millisecond)

	if received != 1 {
		t.Errorf("Expected 1 delivery, got %d", received)
	}
}

func TestWebhookSignature(t *testing.T) {
	var signature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature = r.Header.Get("X-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{
		ID:      "test",
		URL:     server.URL,
		Secret:  "test-secret",
		Enabled: true,
	})

	eb.Publish(NewEvent(EventPlayerJoin))
	time.Sleep(100 * time.Millisecond)

	if signature == "" {
		t.Error("Signature should be set when secret is provided")
	}
	if len(signature) < 10 {
		t.Error("Signature seems too short")
	}
}

func TestWebhookDeliveryHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{
		ID:      "test",
		URL:     server.URL,
		Enabled: true,
	})

	eb.Publish(NewEvent(EventPlayerJoin))
	time.Sleep(100 * time.Millisecond)

	deliveries := wm.GetDeliveries(10)
	if len(deliveries) == 0 {
		t.Error("Should have delivery history")
	}
	if !deliveries[0].Success {
		t.Error("Delivery should be successful")
	}
}

func TestWebhookFailedDelivery(t *testing.T) {
	// Server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{
		ID:         "test",
		URL:        server.URL,
		Enabled:    true,
		RetryCount: 1,
	})

	eb.Publish(NewEvent(EventPlayerJoin))
	time.Sleep(100 * time.Millisecond)

	deliveries := wm.GetDeliveries(10)
	if len(deliveries) == 0 {
		t.Error("Should have delivery history")
	}
	if deliveries[0].Success {
		t.Error("Delivery should have failed")
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "test-secret"

	// Generate signature
	wm := NewWebhookManager(nil, "")
	sig := wm.signPayload(payload, secret)

	// Verify
	if !VerifySignature(payload, sig, secret) {
		t.Error("Signature verification should pass")
	}

	// Wrong secret should fail
	if VerifySignature(payload, sig, "wrong-secret") {
		t.Error("Signature verification should fail with wrong secret")
	}

	// Tampered payload should fail
	if VerifySignature([]byte(`{"test":"tampered"}`), sig, secret) {
		t.Error("Signature verification should fail with tampered payload")
	}
}

func TestGetDeliveriesForWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	wm := NewWebhookManager(eb, "test-server")
	wm.Start()
	defer wm.Stop()

	wm.AddWebhook(&WebhookConfig{ID: "wh1", URL: server.URL, Enabled: true})
	wm.AddWebhook(&WebhookConfig{ID: "wh2", URL: server.URL, Enabled: true})

	eb.Publish(NewEvent(EventPlayerJoin))
	time.Sleep(100 * time.Millisecond)

	deliveries := wm.GetDeliveriesForWebhook("wh1", 10)
	for _, d := range deliveries {
		if d.WebhookID != "wh1" {
			t.Error("Should only return deliveries for wh1")
		}
	}
}
