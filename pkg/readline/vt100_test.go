package readline

import "testing"

func TestParseEscapeSequenceArrowKeys(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		want Key
	}{
		{"Up", []byte{0x1B, '[', 'A'}, KeyUp},
		{"Down", []byte{0x1B, '[', 'B'}, KeyDown},
		{"Right", []byte{0x1B, '[', 'C'}, KeyRight},
		{"Left", []byte{0x1B, '[', 'D'}, KeyLeft},
		{"Home", []byte{0x1B, '[', 'H'}, KeyHome},
		{"End", []byte{0x1B, '[', 'F'}, KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEscapeSequence(tt.buf)
			if got != tt.want {
				t.Errorf("parseEscapeSequence(%v) = %v, want %v", tt.buf, got, tt.want)
			}
		})
	}
}

func TestParseEscapeSequenceSS3(t *testing.T) {
	// Some terminals use ESC O instead of ESC [
	tests := []struct {
		name string
		buf  []byte
		want Key
	}{
		{"Up SS3", []byte{0x1B, 'O', 'A'}, KeyUp},
		{"Down SS3", []byte{0x1B, 'O', 'B'}, KeyDown},
		{"Right SS3", []byte{0x1B, 'O', 'C'}, KeyRight},
		{"Left SS3", []byte{0x1B, 'O', 'D'}, KeyLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEscapeSequence(tt.buf)
			if got != tt.want {
				t.Errorf("parseEscapeSequence(%v) = %v, want %v", tt.buf, got, tt.want)
			}
		})
	}
}

func TestParseEscapeSequenceDelete(t *testing.T) {
	// Delete is ESC [ 3 ~
	buf := []byte{0x1B, '[', '3', '~'}
	got := parseEscapeSequence(buf)
	if got != KeyDelete {
		t.Errorf("Delete sequence = %v, want KeyDelete", got)
	}
}

func TestParseEscapeSequenceIncomplete(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
	}{
		{"Just ESC", []byte{0x1B}},
		{"ESC [", []byte{0x1B, '['}},
		{"Delete incomplete", []byte{0x1B, '[', '3'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEscapeSequence(tt.buf)
			if got != KeyNone {
				t.Errorf("Incomplete sequence should return KeyNone, got %v", got)
			}
		})
	}
}

func TestParseEscapeSequenceInvalid(t *testing.T) {
	// Not starting with ESC
	buf := []byte{'A', 'B', 'C'}
	got := parseEscapeSequence(buf)
	if got != KeyNone {
		t.Errorf("Invalid sequence should return KeyNone, got %v", got)
	}
}

func TestVT100Constants(t *testing.T) {
	// Just verify constants are defined
	if CursorUp == "" {
		t.Error("CursorUp should not be empty")
	}
	if CursorDown == "" {
		t.Error("CursorDown should not be empty")
	}
	if CursorLeft == "" {
		t.Error("CursorLeft should not be empty")
	}
	if CursorRight == "" {
		t.Error("CursorRight should not be empty")
	}
	if ClearToEnd == "" {
		t.Error("ClearToEnd should not be empty")
	}
}

func TestMoveCursor(t *testing.T) {
	// MoveCursor with 0 should return empty
	if MoveCursor(0, KeyLeft) != "" {
		t.Error("MoveCursor(0, ...) should return empty")
	}
	if MoveCursor(-1, KeyLeft) != "" {
		t.Error("MoveCursor(-1, ...) should return empty")
	}

	// Valid movement - all directions
	tests := []struct {
		name      string
		n         int
		direction Key
		wantEmpty bool
	}{
		{"left 1", 1, KeyLeft, false},
		{"right 1", 1, KeyRight, false},
		{"up 1", 1, KeyUp, false},
		{"down 1", 1, KeyDown, false},
		{"left 5", 5, KeyLeft, false},
		{"invalid direction", 1, KeyNone, true},
		{"invalid direction 2", 1, KeyEnter, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MoveCursor(tt.n, tt.direction)
			if tt.wantEmpty && result != "" {
				t.Errorf("MoveCursor(%d, %v) = %q, want empty", tt.n, tt.direction, result)
			}
			if !tt.wantEmpty && result == "" {
				t.Errorf("MoveCursor(%d, %v) should not be empty", tt.n, tt.direction)
			}
		})
	}
}

// TestParseEscapeSequenceHomeEnd tests alternative home/end sequences
func TestParseEscapeSequenceHomeEnd(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		want Key
	}{
		{"Home ESC[1~", []byte{0x1B, '[', '1', '~'}, KeyHome},
		{"End ESC[4~", []byte{0x1B, '[', '4', '~'}, KeyEnd},
		{"Home SS3", []byte{0x1B, 'O', 'H'}, KeyHome},
		{"End SS3", []byte{0x1B, 'O', 'F'}, KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEscapeSequence(tt.buf)
			if got != tt.want {
				t.Errorf("parseEscapeSequence(%v) = %v, want %v", tt.buf, got, tt.want)
			}
		})
	}
}

// TestParseEscapeSequenceUnknown tests unknown sequences return KeyEscape
func TestParseEscapeSequenceUnknown(t *testing.T) {
	// ESC followed by unknown character
	buf := []byte{0x1B, 'X', 'Y'}
	got := parseEscapeSequence(buf)
	if got != KeyEscape {
		t.Errorf("Unknown sequence should return KeyEscape, got %v", got)
	}
}
