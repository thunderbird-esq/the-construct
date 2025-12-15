package readline

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

// editorMockConn is a mock connection for editor testing
type editorMockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func newEditorMockConn(input []byte) *editorMockConn {
	return &editorMockConn{
		readBuf:  bytes.NewBuffer(input),
		writeBuf: &bytes.Buffer{},
	}
}

func (m *editorMockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}
	if m.readBuf.Len() == 0 {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *editorMockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *editorMockConn) Close() error {
	m.closed = true
	return nil
}

func (m *editorMockConn) LocalAddr() net.Addr                { return nil }
func (m *editorMockConn) RemoteAddr() net.Addr               { return nil }
func (m *editorMockConn) SetDeadline(t time.Time) error      { return nil }
func (m *editorMockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *editorMockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestNewEditor verifies Editor creation
func TestNewEditor(t *testing.T) {
	conn := newEditorMockConn(nil)
	history := NewHistory(10)
	
	editor := NewEditor(conn, history)
	
	if editor == nil {
		t.Fatal("NewEditor returned nil")
	}
	if editor.conn != conn {
		t.Error("conn not set")
	}
	if editor.history != history {
		t.Error("history not set")
	}
	if editor.buffer == nil {
		t.Error("buffer should be created")
	}
}

// TestEditorReadLineSimple verifies basic line reading
func TestEditorReadLineSimple(t *testing.T) {
	// Input: "hello" followed by Enter (CR)
	input := []byte{'h', 'e', 'l', 'l', 'o', 0x0D}
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "hello" {
		t.Errorf("ReadLine() = %q, want 'hello'", line)
	}
}

// TestEditorReadLineWithHistory verifies history is recorded
func TestEditorReadLineWithHistory(t *testing.T) {
	input := []byte{'t', 'e', 's', 't', 0x0D}
	conn := newEditorMockConn(input)
	history := NewHistory(10)
	editor := NewEditor(conn, history)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "test" {
		t.Errorf("ReadLine() = %q, want 'test'", line)
	}
	
	// Check history was updated
	if history.Len() != 1 {
		t.Errorf("History length = %d, want 1", history.Len())
	}
	if history.Get(0) != "test" {
		t.Errorf("History[0] = %q, want 'test'", history.Get(0))
	}
}

// TestEditorBackspace verifies backspace handling
func TestEditorBackspace(t *testing.T) {
	// Input: "ab" + backspace + "c" + Enter
	input := []byte{'a', 'b', 0x7F, 'c', 0x0D}
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "ac" {
		t.Errorf("ReadLine() = %q, want 'ac'", line)
	}
}

// TestEditorCtrlC verifies Ctrl+C handling
func TestEditorCtrlC(t *testing.T) {
	input := []byte{'a', 'b', 0x03} // 0x03 = Ctrl+C
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "" {
		t.Errorf("Ctrl+C should return empty line, got %q", line)
	}
}

// TestEditorCtrlD verifies Ctrl+D (EOF) handling
func TestEditorCtrlD(t *testing.T) {
	// Ctrl+D on empty line should return EOF
	input := []byte{0x04} // 0x04 = Ctrl+D
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	_, err := editor.ReadLine()
	
	if err != io.EOF {
		t.Errorf("Ctrl+D on empty line should return EOF, got %v", err)
	}
}

// TestEditorCtrlU verifies Ctrl+U (clear line) handling
func TestEditorCtrlU(t *testing.T) {
	// Input: "abc" + Ctrl+U + "xyz" + Enter
	input := []byte{'a', 'b', 'c', 0x15, 'x', 'y', 'z', 0x0D}
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "xyz" {
		t.Errorf("ReadLine() = %q, want 'xyz'", line)
	}
}

// TestEditorArrowKeys verifies arrow key handling
func TestEditorArrowKeys(t *testing.T) {
	// Input: "ab" + left arrow + "X" + Enter
	// Left arrow is ESC [ D
	input := []byte{'a', 'b', 0x1B, '[', 'D', 'X', 0x0D}
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	// After "ab", cursor at end (pos 2)
	// Left arrow moves cursor to pos 1
	// 'X' inserts at pos 1
	// Result should be "aXb"
	if line != "aXb" {
		t.Errorf("ReadLine() = %q, want 'aXb'", line)
	}
}

// TestEditorHistoryNavigation verifies up/down arrow history
func TestEditorHistoryNavigation(t *testing.T) {
	history := NewHistory(10)
	history.Add("first command")
	history.Add("second command")
	
	// Input: up arrow + Enter (should get "second command")
	input := []byte{0x1B, '[', 'A', 0x0D}
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, history)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "second command" {
		t.Errorf("ReadLine() = %q, want 'second command'", line)
	}
}

// TestEditorEmptyLine verifies empty line not added to history
func TestEditorEmptyLine(t *testing.T) {
	input := []byte{0x0D} // Just Enter
	conn := newEditorMockConn(input)
	history := NewHistory(10)
	editor := NewEditor(conn, history)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "" {
		t.Errorf("Empty line should return empty string, got %q", line)
	}
	if history.Len() != 0 {
		t.Error("Empty line should not be added to history")
	}
}

// TestEditorLFTerminator verifies LF as line terminator
func TestEditorLFTerminator(t *testing.T) {
	input := []byte{'h', 'i', 0x0A} // 0x0A = LF
	conn := newEditorMockConn(input)
	editor := NewEditor(conn, nil)
	
	line, err := editor.ReadLine()
	
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if line != "hi" {
		t.Errorf("ReadLine() = %q, want 'hi'", line)
	}
}
