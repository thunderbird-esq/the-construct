// Package events provides a pub/sub event system for Matrix MUD.
// Enables decoupled communication between game systems and external integrations.
package events

import (
	"encoding/json"
	"sync"
	"time"
)

// EventType represents the type of game event
type EventType string

const (
	// Player events
	EventPlayerJoin       EventType = "player.join"
	EventPlayerLeave      EventType = "player.leave"
	EventPlayerDeath      EventType = "player.death"
	EventPlayerLevelUp    EventType = "player.level_up"
	EventPlayerMove       EventType = "player.move"
	EventPlayerChat       EventType = "player.chat"
	EventPlayerCommand    EventType = "player.command"

	// Combat events
	EventCombatStart      EventType = "combat.start"
	EventCombatEnd        EventType = "combat.end"
	EventCombatHit        EventType = "combat.hit"
	EventCombatMiss       EventType = "combat.miss"
	EventNPCKill          EventType = "combat.npc_kill"
	EventPvPKill          EventType = "combat.pvp_kill"

	// Item events
	EventItemPickup       EventType = "item.pickup"
	EventItemDrop         EventType = "item.drop"
	EventItemEquip        EventType = "item.equip"
	EventItemUnequip      EventType = "item.unequip"
	EventItemCraft        EventType = "item.craft"
	EventItemUse          EventType = "item.use"

	// Quest events
	EventQuestStart       EventType = "quest.start"
	EventQuestProgress    EventType = "quest.progress"
	EventQuestComplete    EventType = "quest.complete"
	EventQuestFail        EventType = "quest.fail"

	// Social events
	EventPartyCreate      EventType = "party.create"
	EventPartyJoin        EventType = "party.join"
	EventPartyLeave       EventType = "party.leave"
	EventPartyDisband     EventType = "party.disband"
	EventFactionJoin      EventType = "faction.join"
	EventFactionLeave     EventType = "faction.leave"
	EventFactionRepChange EventType = "faction.rep_change"

	// Economy events
	EventTradeStart       EventType = "trade.start"
	EventTradeComplete    EventType = "trade.complete"
	EventTradeCancel      EventType = "trade.cancel"
	EventAuctionCreate    EventType = "auction.create"
	EventAuctionBid       EventType = "auction.bid"
	EventAuctionSold      EventType = "auction.sold"
	EventShopBuy          EventType = "shop.buy"
	EventShopSell         EventType = "shop.sell"

	// Achievement events
	EventAchievement      EventType = "achievement.unlock"

	// System events
	EventServerStart      EventType = "server.start"
	EventServerStop       EventType = "server.stop"
	EventServerBroadcast  EventType = "server.broadcast"
	EventAdminAction      EventType = "admin.action"
)

// Event represents a game event
type Event struct {
	ID         string                 `json:"id"`
	Type       EventType              `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	PlayerName string                 `json:"player_name,omitempty"`
	PlayerID   int64                  `json:"player_id,omitempty"`
	RoomID     string                 `json:"room_id,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}
}

// WithPlayer adds player info to the event
func (e *Event) WithPlayer(name string, id int64) *Event {
	e.PlayerName = name
	e.PlayerID = id
	return e
}

// WithRoom adds room info to the event
func (e *Event) WithRoom(roomID string) *Event {
	e.RoomID = roomID
	return e
}

// WithData adds data to the event
func (e *Event) WithData(key string, value interface{}) *Event {
	e.Data[key] = value
	return e
}

// JSON serializes the event to JSON
func (e *Event) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// EventHandler is a function that handles events
type EventHandler func(*Event)

// Subscription represents an event subscription
type Subscription struct {
	ID        string
	EventType EventType
	Handler   EventHandler
	Filter    func(*Event) bool
}

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu            sync.RWMutex
	subscribers   map[EventType][]*Subscription
	allHandlers   []*Subscription
	eventQueue    chan *Event
	workerCount   int
	running       bool
	wg            sync.WaitGroup
	nextSubID     int
}

// NewEventBus creates a new event bus
func NewEventBus(workerCount int) *EventBus {
	if workerCount <= 0 {
		workerCount = 4
	}
	return &EventBus{
		subscribers:  make(map[EventType][]*Subscription),
		allHandlers:  make([]*Subscription, 0),
		eventQueue:   make(chan *Event, 1000),
		workerCount:  workerCount,
		nextSubID:    1,
	}
}

// Start starts the event bus workers
func (eb *EventBus) Start() {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return
	}
	eb.running = true
	eb.mu.Unlock()

	for i := 0; i < eb.workerCount; i++ {
		eb.wg.Add(1)
		go eb.worker()
	}
}

// Stop stops the event bus
func (eb *EventBus) Stop() {
	eb.mu.Lock()
	if !eb.running {
		eb.mu.Unlock()
		return
	}
	eb.running = false
	eb.mu.Unlock()

	close(eb.eventQueue)
	eb.wg.Wait()
}

// worker processes events from the queue
func (eb *EventBus) worker() {
	defer eb.wg.Done()
	for event := range eb.eventQueue {
		eb.dispatch(event)
	}
}

// dispatch sends event to all matching subscribers
func (eb *EventBus) dispatch(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Call type-specific handlers
	if handlers, ok := eb.subscribers[event.Type]; ok {
		for _, sub := range handlers {
			if sub.Filter == nil || sub.Filter(event) {
				sub.Handler(event)
			}
		}
	}

	// Call wildcard handlers
	for _, sub := range eb.allHandlers {
		if sub.Filter == nil || sub.Filter(event) {
			sub.Handler(event)
		}
	}
}

// Subscribe adds a handler for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) string {
	return eb.SubscribeWithFilter(eventType, handler, nil)
}

// SubscribeWithFilter adds a handler with a filter function
func (eb *EventBus) SubscribeWithFilter(eventType EventType, handler EventHandler, filter func(*Event) bool) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subID := generateSubID(eb.nextSubID)
	eb.nextSubID++

	sub := &Subscription{
		ID:        subID,
		EventType: eventType,
		Handler:   handler,
		Filter:    filter,
	}

	eb.subscribers[eventType] = append(eb.subscribers[eventType], sub)
	return subID
}

// SubscribeAll adds a handler for all events
func (eb *EventBus) SubscribeAll(handler EventHandler) string {
	return eb.SubscribeAllWithFilter(handler, nil)
}

// SubscribeAllWithFilter adds a handler for all events with a filter
func (eb *EventBus) SubscribeAllWithFilter(handler EventHandler, filter func(*Event) bool) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subID := generateSubID(eb.nextSubID)
	eb.nextSubID++

	sub := &Subscription{
		ID:      subID,
		Handler: handler,
		Filter:  filter,
	}

	eb.allHandlers = append(eb.allHandlers, sub)
	return subID
}

// Unsubscribe removes a subscription by ID
func (eb *EventBus) Unsubscribe(subID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Check type-specific subscribers
	for eventType, subs := range eb.subscribers {
		for i, sub := range subs {
			if sub.ID == subID {
				eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}

	// Check all-event handlers
	for i, sub := range eb.allHandlers {
		if sub.ID == subID {
			eb.allHandlers = append(eb.allHandlers[:i], eb.allHandlers[i+1:]...)
			return
		}
	}
}

// Publish sends an event to the queue (async)
func (eb *EventBus) Publish(event *Event) {
	eb.mu.RLock()
	running := eb.running
	eb.mu.RUnlock()

	if !running {
		return
	}

	select {
	case eb.eventQueue <- event:
	default:
		// Queue full, drop event (could log this)
	}
}

// PublishSync sends an event and waits for processing
func (eb *EventBus) PublishSync(event *Event) {
	eb.dispatch(event)
}

// QueueSize returns the current queue size
func (eb *EventBus) QueueSize() int {
	return len(eb.eventQueue)
}

// SubscriberCount returns the number of subscribers
func (eb *EventBus) SubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	count := len(eb.allHandlers)
	for _, subs := range eb.subscribers {
		count += len(subs)
	}
	return count
}

// Helper functions
var eventCounter int64
var eventCounterMu sync.Mutex

func generateEventID() string {
	eventCounterMu.Lock()
	defer eventCounterMu.Unlock()
	eventCounter++
	return time.Now().Format("20060102150405") + "-" + itoa(eventCounter)
}

func generateSubID(n int) string {
	return "sub_" + itoa(int64(n))
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// Global event bus instance
var GlobalEventBus = NewEventBus(4)

// Convenience functions using global bus
func Publish(event *Event) {
	GlobalEventBus.Publish(event)
}

func Subscribe(eventType EventType, handler EventHandler) string {
	return GlobalEventBus.Subscribe(eventType, handler)
}

func Unsubscribe(subID string) {
	GlobalEventBus.Unsubscribe(subID)
}
