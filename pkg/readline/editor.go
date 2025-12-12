// Package readline provides line editing with command history for telnet connections.
// It handles escape sequences for arrow keys and provides a readline-like experience.
package readline

import (
	"io"
	"net"
	"time"
)

// Key constants for special keys
type Key int

const (
	KeyNone Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyBackspace
	KeyDelete
	KeyHome
	KeyEnd
	KeyEnter
	KeyCtrlC
	KeyCtrlD
	KeyEscape
	KeyTab
)

// Editor handles line editing with history support
type Editor struct {
	history *History
	buffer  *Buffer
	conn    net.Conn
}

// NewEditor creates a new line editor
func NewEditor(conn net.Conn, history *History) *Editor {
	return &Editor{
		history: history,
		buffer:  NewBuffer(),
		conn:    conn,
	}
}

// ReadLine reads a line with editing support
// Returns the line (without newline) and any error
func (e *Editor) ReadLine() (string, error) {
	e.buffer.Clear()
	historyIndex := -1 // -1 means current input, not in history
	savedInput := ""   // Save current input when browsing history

	buf := make([]byte, 1)
	escBuf := make([]byte, 0, 8)
	inEscape := false

	for {
		// Set read deadline to prevent blocking forever
		e.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := e.conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is OK, just retry
			}
			if err == io.EOF {
				return "", err
			}
			return "", err
		}
		if n == 0 {
			continue
		}

		b := buf[0]

		// Handle escape sequences
		if inEscape {
			escBuf = append(escBuf, b)
			key := parseEscapeSequence(escBuf)
			if key != KeyNone {
				inEscape = false
				escBuf = escBuf[:0]

				switch key {
				case KeyUp:
					if e.history != nil && e.history.Len() > 0 {
						if historyIndex == -1 {
							savedInput = e.buffer.String()
						}
						if historyIndex < e.history.Len()-1 {
							historyIndex++
							e.setLine(e.history.Get(historyIndex))
						}
					}
				case KeyDown:
					if historyIndex > 0 {
						historyIndex--
						e.setLine(e.history.Get(historyIndex))
					} else if historyIndex == 0 {
						historyIndex = -1
						e.setLine(savedInput)
					}
				case KeyLeft:
					if e.buffer.CursorLeft() {
						e.write(CursorLeft)
					}
				case KeyRight:
					if e.buffer.CursorRight() {
						e.write(CursorRight)
					}
				case KeyHome:
					for e.buffer.CursorLeft() {
						e.write(CursorLeft)
					}
				case KeyEnd:
					for e.buffer.CursorRight() {
						e.write(CursorRight)
					}
				case KeyDelete:
					if e.buffer.Delete() {
						e.redrawFromCursor()
					}
				}
				continue
			}
			// Escape sequence not complete yet, or invalid
			if len(escBuf) > 6 {
				// Too long, probably garbage
				inEscape = false
				escBuf = escBuf[:0]
			}
			continue
		}

		switch b {
		case 0x1B: // ESC
			inEscape = true
			escBuf = append(escBuf[:0], b)

		case 0x0D, 0x0A: // CR or LF (Enter)
			e.write("\r\n")
			line := e.buffer.String()
			if line != "" && e.history != nil {
				e.history.Add(line)
			}
			return line, nil

		case 0x7F, 0x08: // DEL or Backspace
			if e.buffer.Backspace() {
				e.write(CursorLeft + " " + CursorLeft)
				e.redrawFromCursor()
			}

		case 0x03: // Ctrl+C
			e.write("^C\r\n")
			return "", nil

		case 0x04: // Ctrl+D (EOF)
			if e.buffer.Len() == 0 {
				return "", io.EOF
			}

		case 0x15: // Ctrl+U (clear line)
			e.clearLine()
			e.buffer.Clear()

		case 0x0C: // Ctrl+L (redraw)
			e.redrawLine()

		default:
			// Regular printable character
			if b >= 0x20 && b < 0x7F {
				e.buffer.Insert(b)
				e.write(string(b))
				e.redrawFromCursor()
			}
		}
	}
}

// setLine replaces the current line with a new one
func (e *Editor) setLine(s string) {
	e.clearLine()
	e.buffer.Clear()
	for _, c := range s {
		e.buffer.Insert(byte(c))
	}
	e.write(s)
}

// clearLine clears the current line on screen
func (e *Editor) clearLine() {
	// Move to start, clear to end
	for i := 0; i < e.buffer.Cursor(); i++ {
		e.write(CursorLeft)
	}
	e.write(ClearToEnd)
}

// redrawLine redraws the entire line
func (e *Editor) redrawLine() {
	e.write("\r> " + e.buffer.String())
	// Move cursor to correct position
	for i := e.buffer.Len(); i > e.buffer.Cursor(); i-- {
		e.write(CursorLeft)
	}
}

// redrawFromCursor redraws from cursor to end
func (e *Editor) redrawFromCursor() {
	// Save cursor position
	pos := e.buffer.Cursor()
	// Write rest of buffer + space to clear any leftover char
	rest := e.buffer.StringFrom(pos)
	e.write(rest + " ")
	// Move back to cursor position
	for i := 0; i <= len(rest); i++ {
		e.write(CursorLeft)
	}
}

func (e *Editor) write(s string) {
	e.conn.Write([]byte(s))
}
