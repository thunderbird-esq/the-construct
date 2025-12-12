package readline

import "testing"

func TestNewHistory(t *testing.T) {
	h := NewHistory(10)
	if h == nil {
		t.Fatal("NewHistory returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("New history should be empty, got %d", h.Len())
	}
}

func TestHistoryAdd(t *testing.T) {
	h := NewHistory(5)

	h.Add("cmd1")
	if h.Len() != 1 {
		t.Errorf("Len = %d, want 1", h.Len())
	}
	if h.Get(0) != "cmd1" {
		t.Errorf("Get(0) = %q, want cmd1", h.Get(0))
	}

	h.Add("cmd2")
	if h.Len() != 2 {
		t.Errorf("Len = %d, want 2", h.Len())
	}
	// Most recent should be at index 0
	if h.Get(0) != "cmd2" {
		t.Errorf("Get(0) = %q, want cmd2", h.Get(0))
	}
	if h.Get(1) != "cmd1" {
		t.Errorf("Get(1) = %q, want cmd1", h.Get(1))
	}
}

func TestHistoryMaxSize(t *testing.T) {
	h := NewHistory(3)

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")
	h.Add("cmd4")

	if h.Len() != 3 {
		t.Errorf("Len = %d, want 3 (max size)", h.Len())
	}
	// Oldest (cmd1) should be gone
	if h.Get(0) != "cmd4" {
		t.Errorf("Get(0) = %q, want cmd4", h.Get(0))
	}
	if h.Get(2) != "cmd2" {
		t.Errorf("Get(2) = %q, want cmd2", h.Get(2))
	}
}

func TestHistoryNoDuplicates(t *testing.T) {
	h := NewHistory(10)

	h.Add("cmd1")
	h.Add("cmd1") // Duplicate

	if h.Len() != 1 {
		t.Errorf("Len = %d, want 1 (no duplicates)", h.Len())
	}
}

func TestHistoryEmpty(t *testing.T) {
	h := NewHistory(10)

	h.Add("") // Empty string should be ignored

	if h.Len() != 0 {
		t.Errorf("Len = %d, want 0 (empty ignored)", h.Len())
	}
}

func TestHistoryGetOutOfBounds(t *testing.T) {
	h := NewHistory(10)
	h.Add("cmd1")

	if h.Get(-1) != "" {
		t.Error("Get(-1) should return empty")
	}
	if h.Get(100) != "" {
		t.Error("Get(100) should return empty")
	}
}

func TestHistoryClear(t *testing.T) {
	h := NewHistory(10)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Clear()

	if h.Len() != 0 {
		t.Errorf("Len = %d after Clear, want 0", h.Len())
	}
}

func TestHistoryAll(t *testing.T) {
	h := NewHistory(10)
	h.Add("cmd1")
	h.Add("cmd2")

	all := h.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d, want 2", len(all))
	}
	if all[0] != "cmd2" || all[1] != "cmd1" {
		t.Errorf("All() = %v, want [cmd2 cmd1]", all)
	}
}
