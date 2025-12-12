package readline

import "testing"

func TestNewBuffer(t *testing.T) {
	b := NewBuffer()
	if b == nil {
		t.Fatal("NewBuffer returned nil")
	}
	if b.Len() != 0 {
		t.Errorf("New buffer should be empty")
	}
	if b.Cursor() != 0 {
		t.Errorf("New buffer cursor should be 0")
	}
}

func TestBufferInsert(t *testing.T) {
	b := NewBuffer()

	b.Insert('a')
	if b.String() != "a" {
		t.Errorf("String = %q, want a", b.String())
	}
	if b.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", b.Cursor())
	}

	b.Insert('b')
	if b.String() != "ab" {
		t.Errorf("String = %q, want ab", b.String())
	}
}

func TestBufferInsertMiddle(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('c')
	b.CursorLeft()
	b.Insert('b')

	if b.String() != "abc" {
		t.Errorf("String = %q, want abc", b.String())
	}
}

func TestBufferBackspace(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('b')

	if !b.Backspace() {
		t.Error("Backspace should return true")
	}
	if b.String() != "a" {
		t.Errorf("String = %q, want a", b.String())
	}

	b.Backspace()
	if b.Backspace() {
		t.Error("Backspace on empty should return false")
	}
}

func TestBufferDelete(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('b')
	b.CursorLeft()
	b.CursorLeft()

	if !b.Delete() {
		t.Error("Delete should return true")
	}
	if b.String() != "b" {
		t.Errorf("String = %q, want b", b.String())
	}
}

func TestBufferDeleteAtEnd(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')

	if b.Delete() {
		t.Error("Delete at end should return false")
	}
}

func TestBufferCursor(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('b')
	b.Insert('c')

	if !b.CursorLeft() {
		t.Error("CursorLeft should return true")
	}
	if b.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", b.Cursor())
	}

	b.CursorLeft()
	b.CursorLeft()
	if b.CursorLeft() {
		t.Error("CursorLeft at start should return false")
	}

	b.CursorRight()
	if b.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", b.Cursor())
	}
}

func TestBufferCursorRightAtEnd(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')

	if b.CursorRight() {
		t.Error("CursorRight at end should return false")
	}
}

func TestBufferClear(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('b')
	b.Clear()

	if b.Len() != 0 {
		t.Errorf("Len = %d after Clear, want 0", b.Len())
	}
	if b.Cursor() != 0 {
		t.Errorf("Cursor = %d after Clear, want 0", b.Cursor())
	}
}

func TestBufferStringFrom(t *testing.T) {
	b := NewBuffer()
	b.Insert('a')
	b.Insert('b')
	b.Insert('c')

	if b.StringFrom(1) != "bc" {
		t.Errorf("StringFrom(1) = %q, want bc", b.StringFrom(1))
	}
	if b.StringFrom(3) != "" {
		t.Errorf("StringFrom(3) = %q, want empty", b.StringFrom(3))
	}
}

func TestBufferSet(t *testing.T) {
	b := NewBuffer()
	b.Set("hello")

	if b.String() != "hello" {
		t.Errorf("String = %q, want hello", b.String())
	}
	if b.Cursor() != 5 {
		t.Errorf("Cursor = %d, want 5", b.Cursor())
	}
}
