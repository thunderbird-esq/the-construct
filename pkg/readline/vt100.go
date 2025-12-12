package readline

// VT100/ANSI escape sequences for terminal control

const (
	// Cursor movement
	CursorUp    = "\x1b[A"
	CursorDown  = "\x1b[B"
	CursorRight = "\x1b[C"
	CursorLeft  = "\x1b[D"
	CursorHome  = "\x1b[H"

	// Line clearing
	ClearToEnd   = "\x1b[K"  // Clear from cursor to end of line
	ClearLine    = "\x1b[2K" // Clear entire line
	ClearScreen  = "\x1b[2J" // Clear entire screen

	// Cursor save/restore
	CursorSave    = "\x1b[s"
	CursorRestore = "\x1b[u"
)

// parseEscapeSequence parses an escape sequence and returns the key
// Returns KeyNone if sequence is incomplete or invalid
func parseEscapeSequence(buf []byte) Key {
	if len(buf) < 2 {
		return KeyNone
	}

	// Must start with ESC
	if buf[0] != 0x1B {
		return KeyNone
	}

	// ESC [ sequences (CSI)
	if buf[1] == '[' {
		if len(buf) < 3 {
			return KeyNone // Incomplete
		}

		switch buf[2] {
		case 'A':
			return KeyUp
		case 'B':
			return KeyDown
		case 'C':
			return KeyRight
		case 'D':
			return KeyLeft
		case 'H':
			return KeyHome
		case 'F':
			return KeyEnd
		case '3':
			// Delete key is ESC [ 3 ~
			if len(buf) >= 4 && buf[3] == '~' {
				return KeyDelete
			}
			return KeyNone // Incomplete
		case '1':
			// Home can also be ESC [ 1 ~
			if len(buf) >= 4 && buf[3] == '~' {
				return KeyHome
			}
			return KeyNone
		case '4':
			// End can also be ESC [ 4 ~
			if len(buf) >= 4 && buf[3] == '~' {
				return KeyEnd
			}
			return KeyNone
		}
	}

	// ESC O sequences (SS3) - some terminals use these for arrow keys
	if buf[1] == 'O' {
		if len(buf) < 3 {
			return KeyNone
		}
		switch buf[2] {
		case 'A':
			return KeyUp
		case 'B':
			return KeyDown
		case 'C':
			return KeyRight
		case 'D':
			return KeyLeft
		case 'H':
			return KeyHome
		case 'F':
			return KeyEnd
		}
	}

	// Single ESC (with timeout would be escape key)
	// For now, treat unrecognized sequences as invalid
	return KeyEscape
}

// MoveCursor returns escape sequence to move cursor n positions
func MoveCursor(n int, direction Key) string {
	if n <= 0 {
		return ""
	}
	switch direction {
	case KeyLeft:
		return "\x1b[" + string(rune('0'+n)) + "D"
	case KeyRight:
		return "\x1b[" + string(rune('0'+n)) + "C"
	case KeyUp:
		return "\x1b[" + string(rune('0'+n)) + "A"
	case KeyDown:
		return "\x1b[" + string(rune('0'+n)) + "B"
	}
	return ""
}
