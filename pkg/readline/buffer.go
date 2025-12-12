package readline

// Buffer handles line editing with cursor position
type Buffer struct {
	data   []byte
	cursor int
}

// NewBuffer creates a new empty buffer
func NewBuffer() *Buffer {
	return &Buffer{
		data:   make([]byte, 0, 256),
		cursor: 0,
	}
}

// Insert inserts a character at the cursor position
func (b *Buffer) Insert(c byte) {
	if b.cursor == len(b.data) {
		b.data = append(b.data, c)
	} else {
		b.data = append(b.data[:b.cursor+1], b.data[b.cursor:]...)
		b.data[b.cursor] = c
	}
	b.cursor++
}

// Backspace deletes the character before the cursor
// Returns true if a character was deleted
func (b *Buffer) Backspace() bool {
	if b.cursor == 0 {
		return false
	}
	b.cursor--
	b.data = append(b.data[:b.cursor], b.data[b.cursor+1:]...)
	return true
}

// Delete deletes the character at the cursor
// Returns true if a character was deleted
func (b *Buffer) Delete() bool {
	if b.cursor >= len(b.data) {
		return false
	}
	b.data = append(b.data[:b.cursor], b.data[b.cursor+1:]...)
	return true
}

// CursorLeft moves cursor left, returns true if moved
func (b *Buffer) CursorLeft() bool {
	if b.cursor > 0 {
		b.cursor--
		return true
	}
	return false
}

// CursorRight moves cursor right, returns true if moved
func (b *Buffer) CursorRight() bool {
	if b.cursor < len(b.data) {
		b.cursor++
		return true
	}
	return false
}

// Cursor returns the current cursor position
func (b *Buffer) Cursor() int {
	return b.cursor
}

// Len returns the length of the buffer
func (b *Buffer) Len() int {
	return len(b.data)
}

// String returns the buffer contents as a string
func (b *Buffer) String() string {
	return string(b.data)
}

// StringFrom returns buffer contents from position to end
func (b *Buffer) StringFrom(pos int) string {
	if pos >= len(b.data) {
		return ""
	}
	return string(b.data[pos:])
}

// Clear clears the buffer
func (b *Buffer) Clear() {
	b.data = b.data[:0]
	b.cursor = 0
}

// Set sets the buffer to a string
func (b *Buffer) Set(s string) {
	b.data = []byte(s)
	b.cursor = len(b.data)
}
