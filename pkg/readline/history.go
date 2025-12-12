package readline

// History stores command history with a ring buffer
type History struct {
	commands []string
	maxSize  int
}

// NewHistory creates a new history with the given max size
func NewHistory(maxSize int) *History {
	return &History{
		commands: make([]string, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add adds a command to history (most recent at index 0)
func (h *History) Add(cmd string) {
	if cmd == "" {
		return
	}
	// Don't add duplicates of the most recent command
	if len(h.commands) > 0 && h.commands[0] == cmd {
		return
	}
	// Prepend to front
	h.commands = append([]string{cmd}, h.commands...)
	// Trim if over max
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[:h.maxSize]
	}
}

// Get returns the command at index (0 = most recent)
func (h *History) Get(index int) string {
	if index < 0 || index >= len(h.commands) {
		return ""
	}
	return h.commands[index]
}

// Len returns the number of commands in history
func (h *History) Len() int {
	return len(h.commands)
}

// Clear clears all history
func (h *History) Clear() {
	h.commands = h.commands[:0]
}

// All returns all commands (most recent first)
func (h *History) All() []string {
	result := make([]string, len(h.commands))
	copy(result, h.commands)
	return result
}
