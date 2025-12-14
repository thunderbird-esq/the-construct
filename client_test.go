package main

import (
	"bufio"
	"bytes"
	"net"
	"strings"
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
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *mockConn) output() string {
	return m.writeBuf.String()
}

// TestClientWrite verifies Client.Write method
func TestClientWrite(t *testing.T) {
	conn := newMockConn("")
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	client.Write("Hello, World!")
	
	if conn.output() != "Hello, World!" {
		t.Errorf("Write output = %q, want %q", conn.output(), "Hello, World!")
	}
}

// TestClientSuppressEcho verifies echo suppression
func TestClientSuppressEcho(t *testing.T) {
	conn := newMockConn("")
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	client.suppressEcho()
	
	output := conn.writeBuf.Bytes()
	// Should send IAC WILL ECHO (255, 251, 1)
	if len(output) != 3 || output[0] != TelnetIAC || output[1] != TelnetWILL || output[2] != TelnetECHO {
		t.Errorf("suppressEcho sent %v, want [255 251 1]", output)
	}
}

// TestClientResumeEcho verifies echo resumption
func TestClientResumeEcho(t *testing.T) {
	conn := newMockConn("")
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	client.resumeEcho()
	
	output := conn.writeBuf.Bytes()
	// Should send IAC WONT ECHO (255, 252, 1)
	if len(output) != 3 || output[0] != TelnetIAC || output[1] != TelnetWONT || output[2] != TelnetECHO {
		t.Errorf("resumeEcho sent %v, want [255 252 1]", output)
	}
}

// TestClientReadPassword verifies password reading
func TestClientReadPassword(t *testing.T) {
	conn := newMockConn("secretpassword\n")
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	password, err := client.readPassword()
	if err != nil {
		t.Fatalf("readPassword error: %v", err)
	}
	
	if password != "secretpassword" {
		t.Errorf("password = %q, want %q", password, "secretpassword")
	}

	// Should have sent IAC commands
	output := conn.writeBuf.Bytes()
	if len(output) < 6 {
		t.Error("readPassword should send echo suppression commands")
	}
}

// TestChooseClass verifies class selection
func TestChooseClass(t *testing.T) {
	tests := []struct {
		input     string
		wantClass string
		wantHP    int
	}{
		{"1\n", "Hacker", 15},
		{"2\n", "Rebel", 30},
		{"3\n", "Operator", 20},
	}

	for _, tt := range tests {
		conn := newMockConn(tt.input)
		client := &Client{
			conn:   conn,
			reader: bufio.NewReader(conn),
		}

		player := &Player{
			Name:      "TestPlayer",
			Inventory: []*Item{},
		}

		chooseClass(client, player)

		if player.Class != tt.wantClass {
			t.Errorf("input %q: Class = %q, want %q", tt.input, player.Class, tt.wantClass)
		}
		if player.HP != tt.wantHP {
			t.Errorf("input %q: HP = %d, want %d", tt.input, player.HP, tt.wantHP)
		}
		if len(player.Inventory) == 0 {
			t.Errorf("input %q: should have starting item", tt.input)
		}
	}
}

// TestChooseClassInvalidThenValid verifies invalid input handling
func TestChooseClassInvalidThenValid(t *testing.T) {
	conn := newMockConn("x\n5\n1\n")
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{},
	}

	chooseClass(client, player)

	if player.Class != "Hacker" {
		t.Errorf("Class = %q, want Hacker after invalid inputs", player.Class)
	}

	// Should have prompted for invalid choices
	output := conn.output()
	if !strings.Contains(output, "Invalid") {
		t.Error("Should show 'Invalid' for bad input")
	}
}

// TestBroadcastWithPlayers verifies broadcast to players in room
func TestBroadcastWithPlayers(t *testing.T) {
	world := NewWorld()

	// Create mock clients
	conn1 := newMockConn("")
	client1 := &Client{conn: conn1, reader: bufio.NewReader(conn1)}
	
	conn2 := newMockConn("")
	client2 := &Client{conn: conn2, reader: bufio.NewReader(conn2)}
	
	conn3 := newMockConn("")
	client3 := &Client{conn: conn3, reader: bufio.NewReader(conn3)}

	player1 := &Player{Name: "Player1", RoomID: "dojo", Conn: client1}
	player2 := &Player{Name: "Player2", RoomID: "dojo", Conn: client2}
	player3 := &Player{Name: "Player3", RoomID: "city_1", Conn: client3}

	world.Players[client1] = player1
	world.Players[client2] = player2
	world.Players[client3] = player3

	// Broadcast to dojo, excluding player1
	world.Broadcast("dojo", player1, "Test message")

	// Player2 should receive (same room, not excluded)
	if !strings.Contains(conn2.output(), "Test message") {
		t.Error("Player2 should receive broadcast")
	}

	// Player1 should NOT receive (excluded)
	if strings.Contains(conn1.output(), "Test message") {
		t.Error("Player1 should NOT receive broadcast (excluded)")
	}

	// Player3 should NOT receive (different room)
	if strings.Contains(conn3.output(), "Test message") {
		t.Error("Player3 should NOT receive broadcast (different room)")
	}
}
