package events

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventPlayerJoin)
	
	if event.Type != EventPlayerJoin {
		t.Errorf("Type = %s, want player.join", event.Type)
	}
	if event.ID == "" {
		t.Error("ID should be set")
	}
	if event.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
	if event.Data == nil {
		t.Error("Data should be initialized")
	}
}

func TestEventChaining(t *testing.T) {
	event := NewEvent(EventPlayerMove).
		WithPlayer("TestPlayer", 123).
		WithRoom("dojo").
		WithData("direction", "north")

	if event.PlayerName != "TestPlayer" {
		t.Errorf("PlayerName = %s, want TestPlayer", event.PlayerName)
	}
	if event.PlayerID != 123 {
		t.Errorf("PlayerID = %d, want 123", event.PlayerID)
	}
	if event.RoomID != "dojo" {
		t.Errorf("RoomID = %s, want dojo", event.RoomID)
	}
	if event.Data["direction"] != "north" {
		t.Error("Data should contain direction")
	}
}

func TestEventJSON(t *testing.T) {
	event := NewEvent(EventPlayerJoin).
		WithPlayer("Test", 1).
		WithData("key", "value")

	data, err := event.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("JSON should not be empty")
	}
}

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus(4)
	if eb == nil {
		t.Fatal("NewEventBus returned nil")
	}
	if eb.workerCount != 4 {
		t.Errorf("workerCount = %d, want 4", eb.workerCount)
	}
}

func TestEventBusStartStop(t *testing.T) {
	eb := NewEventBus(2)
	
	eb.Start()
	// Should be idempotent
	eb.Start()
	
	eb.Stop()
	// Should be idempotent
	eb.Stop()
}

func TestEventBusSubscribe(t *testing.T) {
	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	var received int32
	subID := eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
	})

	if subID == "" {
		t.Error("Subscribe should return subscription ID")
	}
	if eb.SubscriberCount() != 1 {
		t.Errorf("SubscriberCount = %d, want 1", eb.SubscriberCount())
	}
}

func TestEventBusPublish(t *testing.T) {
	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	var received int32
	var wg sync.WaitGroup
	wg.Add(1)

	eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
		wg.Done()
	})

	eb.Publish(NewEvent(EventPlayerJoin))

	// Wait for event to be processed
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if received != 1 {
			t.Errorf("Expected 1 event, got %d", received)
		}
	case <-time.After(time.Second):
		t.Error("Event not received within timeout")
	}
}

func TestEventBusPublishSync(t *testing.T) {
	eb := NewEventBus(2)
	// Don't start workers - PublishSync should still work

	var received int32
	eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
	})

	eb.PublishSync(NewEvent(EventPlayerJoin))

	if received != 1 {
		t.Errorf("Expected 1 event, got %d", received)
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	eb := NewEventBus(2)

	var received int32
	subID := eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
	})

	eb.Unsubscribe(subID)

	if eb.SubscriberCount() != 0 {
		t.Errorf("SubscriberCount = %d, want 0", eb.SubscriberCount())
	}

	eb.PublishSync(NewEvent(EventPlayerJoin))
	if received != 0 {
		t.Error("Unsubscribed handler should not receive events")
	}
}

func TestEventBusSubscribeAll(t *testing.T) {
	eb := NewEventBus(2)

	var received int32
	eb.SubscribeAll(func(e *Event) {
		atomic.AddInt32(&received, 1)
	})

	eb.PublishSync(NewEvent(EventPlayerJoin))
	eb.PublishSync(NewEvent(EventPlayerLeave))
	eb.PublishSync(NewEvent(EventNPCKill))

	if received != 3 {
		t.Errorf("Expected 3 events, got %d", received)
	}
}

func TestEventBusFilter(t *testing.T) {
	eb := NewEventBus(2)

	var received int32
	eb.SubscribeWithFilter(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
	}, func(e *Event) bool {
		return e.PlayerName == "SpecificPlayer"
	})

	eb.PublishSync(NewEvent(EventPlayerJoin).WithPlayer("OtherPlayer", 1))
	eb.PublishSync(NewEvent(EventPlayerJoin).WithPlayer("SpecificPlayer", 2))

	if received != 1 {
		t.Errorf("Expected 1 filtered event, got %d", received)
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	eb := NewEventBus(2)

	var count1, count2 int32
	eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&count1, 1)
	})
	eb.Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&count2, 1)
	})

	eb.PublishSync(NewEvent(EventPlayerJoin))

	if count1 != 1 || count2 != 1 {
		t.Errorf("Both subscribers should receive event")
	}
}

func TestEventBusQueueSize(t *testing.T) {
	eb := NewEventBus(2)
	// Don't start - events will queue

	for i := 0; i < 10; i++ {
		eb.Publish(NewEvent(EventPlayerJoin))
	}

	// Events dropped because bus not running
	if eb.QueueSize() != 0 {
		t.Errorf("Queue should be empty when bus not running")
	}

	eb.Start()
	for i := 0; i < 10; i++ {
		eb.Publish(NewEvent(EventPlayerJoin))
	}
	eb.Stop()
}

func TestEventTypes(t *testing.T) {
	// Verify event type constants are defined
	types := []EventType{
		EventPlayerJoin, EventPlayerLeave, EventPlayerDeath, EventPlayerLevelUp,
		EventCombatStart, EventCombatEnd, EventNPCKill, EventPvPKill,
		EventItemPickup, EventItemDrop, EventItemCraft,
		EventQuestStart, EventQuestComplete,
		EventPartyCreate, EventPartyJoin,
		EventTradeStart, EventTradeComplete,
		EventAchievement,
		EventServerStart, EventServerStop,
	}

	for _, et := range types {
		if et == "" {
			t.Error("Event type should not be empty")
		}
	}
}

func TestGlobalEventBus(t *testing.T) {
	if GlobalEventBus == nil {
		t.Error("GlobalEventBus should be initialized")
	}
}

func TestGlobalConvenienceFunctions(t *testing.T) {
	var received int32
	subID := Subscribe(EventPlayerJoin, func(e *Event) {
		atomic.AddInt32(&received, 1)
	})

	if subID == "" {
		t.Error("Subscribe should return ID")
	}

	Unsubscribe(subID)
}

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	id2 := generateEventID()

	if id1 == "" {
		t.Error("Event ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Event IDs should be unique")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{999999, "999999"},
	}

	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
