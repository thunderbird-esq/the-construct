// Package events provides webhook integration for external notifications.
package events

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebhookConfig defines a webhook endpoint configuration
type WebhookConfig struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	URL        string      `json:"url"`
	Secret     string      `json:"secret,omitempty"`
	Events     []EventType `json:"events"`
	Enabled    bool        `json:"enabled"`
	RetryCount int         `json:"retry_count"`
	Timeout    int         `json:"timeout_seconds"`
}

// WebhookPayload is the payload sent to webhooks
type WebhookPayload struct {
	Event     *Event    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
	ServerID  string    `json:"server_id"`
}

// WebhookDelivery tracks a webhook delivery attempt
type WebhookDelivery struct {
	ID           string        `json:"id"`
	WebhookID    string        `json:"webhook_id"`
	EventID      string        `json:"event_id"`
	URL          string        `json:"url"`
	RequestBody  string        `json:"request_body,omitempty"`
	ResponseCode int           `json:"response_code"`
	ResponseBody string        `json:"response_body,omitempty"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"duration_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	Attempt      int           `json:"attempt"`
}

// WebhookManager manages webhook subscriptions and delivery
type WebhookManager struct {
	mu                 sync.RWMutex
	webhooks           map[string]*WebhookConfig
	deliveries         []*WebhookDelivery
	eventBus           *EventBus
	serverID           string
	httpClient         *http.Client
	subID              string
	maxDeliveryHistory int
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(eventBus *EventBus, serverID string) *WebhookManager {
	return &WebhookManager{
		webhooks:           make(map[string]*WebhookConfig),
		deliveries:         make([]*WebhookDelivery, 0),
		eventBus:           eventBus,
		serverID:           serverID,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		maxDeliveryHistory: 1000,
	}
}

// Start subscribes to events and begins processing
func (wm *WebhookManager) Start() {
	wm.subID = wm.eventBus.SubscribeAll(wm.handleEvent)
}

// Stop unsubscribes from events
func (wm *WebhookManager) Stop() {
	if wm.subID != "" {
		wm.eventBus.Unsubscribe(wm.subID)
	}
}

// AddWebhook adds a new webhook configuration
func (wm *WebhookManager) AddWebhook(config *WebhookConfig) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if config.ID == "" {
		return fmt.Errorf("webhook ID is required")
	}
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}

	wm.webhooks[config.ID] = config
	return nil
}

// RemoveWebhook removes a webhook by ID
func (wm *WebhookManager) RemoveWebhook(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.webhooks, id)
}

// GetWebhook returns a webhook by ID
func (wm *WebhookManager) GetWebhook(id string) *WebhookConfig {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.webhooks[id]
}

// ListWebhooks returns all webhooks
func (wm *WebhookManager) ListWebhooks() []*WebhookConfig {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhooks := make([]*WebhookConfig, 0, len(wm.webhooks))
	for _, wh := range wm.webhooks {
		webhooks = append(webhooks, wh)
	}
	return webhooks
}

// EnableWebhook enables a webhook
func (wm *WebhookManager) EnableWebhook(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wh, ok := wm.webhooks[id]
	if !ok {
		return fmt.Errorf("webhook not found: %s", id)
	}
	wh.Enabled = true
	return nil
}

// DisableWebhook disables a webhook
func (wm *WebhookManager) DisableWebhook(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wh, ok := wm.webhooks[id]
	if !ok {
		return fmt.Errorf("webhook not found: %s", id)
	}
	wh.Enabled = false
	return nil
}

// handleEvent processes events and delivers to webhooks
func (wm *WebhookManager) handleEvent(event *Event) {
	wm.mu.RLock()
	webhooks := make([]*WebhookConfig, 0)
	for _, wh := range wm.webhooks {
		if wh.Enabled && wm.shouldDeliver(wh, event) {
			webhooks = append(webhooks, wh)
		}
	}
	wm.mu.RUnlock()

	for _, wh := range webhooks {
		go wm.deliver(wh, event)
	}
}

// shouldDeliver checks if webhook should receive the event
func (wm *WebhookManager) shouldDeliver(wh *WebhookConfig, event *Event) bool {
	if len(wh.Events) == 0 {
		return true // No filter, deliver all
	}
	for _, et := range wh.Events {
		if et == event.Type {
			return true
		}
	}
	return false
}

// deliver sends the event to a webhook
func (wm *WebhookManager) deliver(wh *WebhookConfig, event *Event) {
	payload := &WebhookPayload{
		Event:     event,
		Timestamp: time.Now(),
		ServerID:  wm.serverID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		wm.recordDelivery(wh, event, nil, 0, err, 1)
		return
	}

	for attempt := 1; attempt <= wh.RetryCount; attempt++ {
		delivery := wm.attemptDelivery(wh, event, body, attempt)
		if delivery.Success {
			break
		}
		if attempt < wh.RetryCount {
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}
	}
}

// attemptDelivery makes a single delivery attempt
func (wm *WebhookManager) attemptDelivery(wh *WebhookConfig, event *Event, body []byte, attempt int) *WebhookDelivery {
	start := time.Now()

	req, err := http.NewRequest("POST", wh.URL, bytes.NewReader(body))
	if err != nil {
		return wm.recordDelivery(wh, event, body, 0, err, attempt)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MatrixMUD-Webhook/1.0")
	req.Header.Set("X-Event-Type", string(event.Type))
	req.Header.Set("X-Event-ID", event.ID)
	req.Header.Set("X-Delivery-Attempt", itoa(int64(attempt)))

	// Add HMAC signature if secret is set
	if wh.Secret != "" {
		sig := wm.signPayload(body, wh.Secret)
		req.Header.Set("X-Signature", sig)
	}

	resp, err := wm.httpClient.Do(req)
	if err != nil {
		return wm.recordDelivery(wh, event, body, 0, err, attempt)
	}
	defer resp.Body.Close()

	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)

	duration := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	delivery := &WebhookDelivery{
		ID:           generateEventID(),
		WebhookID:    wh.ID,
		EventID:      event.ID,
		URL:          wh.URL,
		RequestBody:  string(body),
		ResponseCode: resp.StatusCode,
		ResponseBody: respBody.String(),
		Success:      success,
		Duration:     duration,
		Timestamp:    start,
		Attempt:      attempt,
	}

	if !success {
		delivery.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	wm.storeDelivery(delivery)
	return delivery
}

// recordDelivery records a failed delivery
func (wm *WebhookManager) recordDelivery(wh *WebhookConfig, event *Event, body []byte, code int, err error, attempt int) *WebhookDelivery {
	delivery := &WebhookDelivery{
		ID:           generateEventID(),
		WebhookID:    wh.ID,
		EventID:      event.ID,
		URL:          wh.URL,
		RequestBody:  string(body),
		ResponseCode: code,
		Success:      false,
		Error:        err.Error(),
		Timestamp:    time.Now(),
		Attempt:      attempt,
	}
	wm.storeDelivery(delivery)
	return delivery
}

// storeDelivery stores a delivery record
func (wm *WebhookManager) storeDelivery(delivery *WebhookDelivery) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.deliveries = append(wm.deliveries, delivery)

	// Trim old deliveries
	if len(wm.deliveries) > wm.maxDeliveryHistory {
		wm.deliveries = wm.deliveries[len(wm.deliveries)-wm.maxDeliveryHistory:]
	}
}

// GetDeliveries returns recent deliveries
func (wm *WebhookManager) GetDeliveries(limit int) []*WebhookDelivery {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if limit <= 0 || limit > len(wm.deliveries) {
		limit = len(wm.deliveries)
	}

	start := len(wm.deliveries) - limit
	result := make([]*WebhookDelivery, limit)
	copy(result, wm.deliveries[start:])
	return result
}

// GetDeliveriesForWebhook returns deliveries for a specific webhook
func (wm *WebhookManager) GetDeliveriesForWebhook(webhookID string, limit int) []*WebhookDelivery {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	result := make([]*WebhookDelivery, 0)
	for i := len(wm.deliveries) - 1; i >= 0 && len(result) < limit; i-- {
		if wm.deliveries[i].WebhookID == webhookID {
			result = append(result, wm.deliveries[i])
		}
	}
	return result
}

// signPayload creates HMAC-SHA256 signature
func (wm *WebhookManager) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies a webhook signature
func VerifySignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
