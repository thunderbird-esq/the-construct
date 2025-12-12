package readline

import (
	"bufio"
	"io"
	"net"
	"strings"
	"time"
)

// Reader wraps a connection with readline capabilities
type Reader struct {
	conn     net.Conn
	history  *History
	prompt   string
	editor   *Editor
	fallback *bufio.Reader
}

// NewReader creates a new readline-capable reader
func NewReader(conn net.Conn, history *History, prompt string) *Reader {
	return &Reader{
		conn:     conn,
		history:  history,
		prompt:   prompt,
		editor:   NewEditor(conn, history),
		fallback: bufio.NewReader(conn),
	}
}

// ReadLine reads a line with optional readline support
// Falls back to simple line reading if escape sequences aren't working
func (r *Reader) ReadLine() (string, error) {
	// Try using the editor for full readline support
	line, err := r.editor.ReadLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// ReadLineSimple reads a line without readline support (fallback)
func (r *Reader) ReadLineSimple() (string, error) {
	line, err := r.fallback.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// SimpleReader provides basic line reading without escape sequence handling
// Use this for login/password input where readline features aren't needed
type SimpleReader struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewSimpleReader creates a reader for basic line input
func NewSimpleReader(conn net.Conn) *SimpleReader {
	return &SimpleReader{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// ReadLine reads a single line
func (r *SimpleReader) ReadLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// ReadPassword reads a line with echo suppression
func (r *SimpleReader) ReadPassword() (string, error) {
	// Send IAC WILL ECHO to suppress client echo
	r.conn.Write([]byte{255, 251, 1})
	defer func() {
		// Send IAC WONT ECHO to resume
		r.conn.Write([]byte{255, 252, 1})
		r.conn.Write([]byte("\r\n"))
	}()

	return r.ReadLine()
}

// SetDeadline sets the read deadline on the connection
func (r *Reader) SetDeadline(t time.Time) error {
	return r.conn.SetDeadline(t)
}

// SetReadDeadline sets just the read deadline
func (r *Reader) SetReadDeadline(t time.Time) error {
	return r.conn.SetReadDeadline(t)
}

// Buffered returns bytes buffered in the underlying reader
func (r *Reader) Buffered() int {
	return r.fallback.Buffered()
}

// Read implements io.Reader for compatibility
func (r *Reader) Read(p []byte) (n int, err error) {
	return r.fallback.Read(p)
}

var _ io.Reader = (*Reader)(nil)
