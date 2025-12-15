package readline

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func newMockConn(input string) *mockConn {
	return &mockConn{
		readBuf:  bytes.NewBufferString(input),
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestNewReader verifies Reader creation
func TestNewReader(t *testing.T) {
	conn := newMockConn("")
	history := NewHistory(10)

	reader := NewReader(conn, history, "> ")

	if reader == nil {
		t.Fatal("NewReader returned nil")
	}
	if reader.conn != conn {
		t.Error("conn not set correctly")
	}
	if reader.history != history {
		t.Error("history not set correctly")
	}
	if reader.prompt != "> " {
		t.Errorf("prompt = %q, want '> '", reader.prompt)
	}
	if reader.editor == nil {
		t.Error("editor should be created")
	}
	if reader.fallback == nil {
		t.Error("fallback reader should be created")
	}
}

// TestNewSimpleReader verifies SimpleReader creation
func TestNewSimpleReader(t *testing.T) {
	conn := newMockConn("")

	reader := NewSimpleReader(conn)

	if reader == nil {
		t.Fatal("NewSimpleReader returned nil")
	}
	if reader.conn != conn {
		t.Error("conn not set correctly")
	}
	if reader.reader == nil {
		t.Error("reader should be created")
	}
}

// TestSimpleReaderReadLine verifies basic line reading
func TestSimpleReaderReadLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple line", "hello\n", "hello", false},
		{"with spaces", "  test  \n", "test", false},
		{"empty line", "\n", "", false},
		{"multiple words", "one two three\n", "one two three", false},
		{"with carriage return", "test\r\n", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newMockConn(tt.input)
			reader := NewSimpleReader(conn)

			got, err := reader.ReadLine()

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSimpleReaderReadPassword verifies password reading with echo suppression
func TestSimpleReaderReadPassword(t *testing.T) {
	conn := newMockConn("secret123\n")
	reader := NewSimpleReader(conn)

	pass, err := reader.ReadPassword()

	if err != nil {
		t.Fatalf("ReadPassword() error = %v", err)
	}
	if pass != "secret123" {
		t.Errorf("ReadPassword() = %q, want 'secret123'", pass)
	}

	// Check that IAC sequences were sent
	written := conn.writeBuf.Bytes()
	// Should contain IAC WILL ECHO (255, 251, 1) at start
	if len(written) < 3 || written[0] != 255 || written[1] != 251 || written[2] != 1 {
		t.Error("Should send IAC WILL ECHO before password")
	}
	// Should contain IAC WONT ECHO (255, 252, 1) at end
	found := false
	for i := 3; i < len(written)-2; i++ {
		if written[i] == 255 && written[i+1] == 252 && written[i+2] == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should send IAC WONT ECHO after password")
	}
}

// TestReaderSetDeadline verifies deadline setting
func TestReaderSetDeadline(t *testing.T) {
	conn := newMockConn("")
	reader := NewReader(conn, nil, "> ")

	err := reader.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetDeadline() error = %v", err)
	}

	err = reader.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetReadDeadline() error = %v", err)
	}
}

// TestReaderBuffered verifies buffered bytes reporting
func TestReaderBuffered(t *testing.T) {
	conn := newMockConn("some data")
	reader := NewReader(conn, nil, "> ")

	// Read some data to buffer it
	buf := make([]byte, 4)
	reader.Read(buf)

	// Check buffered
	buffered := reader.Buffered()
	t.Logf("Buffered bytes: %d", buffered)
}

// TestReaderRead verifies io.Reader implementation
func TestReaderRead(t *testing.T) {
	conn := newMockConn("test data")
	reader := NewReader(conn, nil, "> ")

	buf := make([]byte, 4)
	n, err := reader.Read(buf)

	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if n != 4 {
		t.Errorf("Read() n = %d, want 4", n)
	}
	if string(buf) != "test" {
		t.Errorf("Read() = %q, want 'test'", string(buf))
	}
}

// TestReaderReadLineSimple verifies fallback reading
func TestReaderReadLineSimple(t *testing.T) {
	conn := newMockConn("simple line\n")
	reader := NewReader(conn, nil, "> ")

	line, err := reader.ReadLineSimple()

	if err != nil {
		t.Errorf("ReadLineSimple() error = %v", err)
	}
	if line != "simple line" {
		t.Errorf("ReadLineSimple() = %q, want 'simple line'", line)
	}
}
