package validation

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"valid simple", "player1", true},
		{"valid underscore", "player_one", true},
		{"valid mixed case", "PlayerOne", true},
		{"too short", "ab", false},
		{"too long", "abcdefghijklmnopqrstuvwxyz", false},
		{"special chars", "player@name", false},
		{"spaces", "player name", false},
		{"empty", "", false},
		{"min length", "abc", true},
		{"max length", "abcdefghijklmnopqrst", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUsername(tt.username); got != tt.want {
				t.Errorf("ValidateUsername(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"simple", "look", true},
		{"with space", "go north", true},
		{"uppercase", "LOOK", true}, // Should be converted to lowercase
		{"numbers", "look123", false},
		{"special", "look!", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCommand(tt.cmd); got != tt.want {
				t.Errorf("ValidateCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestValidateRoomID(t *testing.T) {
	tests := []struct {
		name   string
		roomID string
		want   bool
	}{
		{"simple", "dojo", true},
		{"with underscore", "city_12", true},
		{"with hyphen", "room-1", true},
		{"mixed", "City_Room-1", true},
		{"too long", strings.Repeat("a", 51), false},
		{"special chars", "room@1", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateRoomID(tt.roomID); got != tt.want {
				t.Errorf("ValidateRoomID(%q) = %v, want %v", tt.roomID, got, tt.want)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal text", "hello world", "hello world"},
		{"with newline", "hello\nworld", "hello\nworld"},
		{"with tab", "hello\tworld", "hello\tworld"},
		{"control chars", "hello\x00world", "helloworld"},
		{"leading space", "  hello", "hello"},
		{"trailing space", "hello  ", "hello"},
		{"escape sequence", "hello\x1b[31mworld", "hello[31mworld"}, // Only removes ESC char, not full sequence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeInput(tt.input); got != tt.want {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"lowercase", "player", "player"},
		{"uppercase", "PLAYER", "player"},
		{"mixed", "Player", "player"},
		{"with spaces", "  player  ", "player"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeUsername(tt.username); got != tt.want {
				t.Errorf("SanitizeUsername(%q) = %q, want %q", tt.username, got, tt.want)
			}
		})
	}
}

func TestIsPrintable(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"normal", "hello world", true},
		{"with newline", "hello\nworld", true},
		{"with tab", "hello\tworld", true},
		{"with null", "hello\x00world", false},
		{"unicode", "héllo wörld", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPrintable(tt.s); got != tt.want {
				t.Errorf("IsPrintable(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOk   bool
	}{
		{"valid", "password123", true},
		{"too short", "pass", false},
		{"exactly 8", "12345678", true},
		{"long", "verylongpassword123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidatePasswordStrength(tt.password)
			if ok != tt.wantOk {
				t.Errorf("ValidatePasswordStrength(%q) = %v, want %v", tt.password, ok, tt.wantOk)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"no truncate", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 5, "hello"},
		{"empty", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.s, tt.maxLen); got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestValidateItemName(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		want     bool
	}{
		{"simple", "sword", true},
		{"with space", "long sword", true},
		{"with underscore", "health_potion", true},
		{"empty", "", false},
		{"too long", strings.Repeat("a", 51), false},
		{"special chars", "sword@1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateItemName(tt.itemName); got != tt.want {
				t.Errorf("ValidateItemName(%q) = %v, want %v", tt.itemName, got, tt.want)
			}
		})
	}
}

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		wantOk bool
	}{
		{"valid", "Hello everyone!", true},
		{"empty", "", false},
		{"too long", strings.Repeat("a", 501), false},
		{"with control char", "hello\x00world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateMessage(tt.msg)
			if ok != tt.wantOk {
				t.Errorf("ValidateMessage(%q) = %v, want %v", tt.msg, ok, tt.wantOk)
			}
		})
	}
}

func TestValidateDirection(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"north", "north", true},
		{"n", "n", true},
		{"south", "south", true},
		{"uppercase", "NORTH", true},
		{"invalid", "sideways", false},
		{"down alias", "dn", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateDirection(tt.dir); got != tt.want {
				t.Errorf("ValidateDirection(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		name string
		qty  int
		want bool
	}{
		{"valid", 5, true},
		{"one", 1, true},
		{"max", 999, true},
		{"zero", 0, false},
		{"negative", -1, false},
		{"too high", 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateQuantity(tt.qty); got != tt.want {
				t.Errorf("ValidateQuantity(%d) = %v, want %v", tt.qty, got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"no ansi", "hello", "hello"},
		{"red text", "\x1b[31mhello\x1b[0m", "hello"},
		{"bold", "\x1b[1mhello\x1b[0m", "hello"},
		{"multiple", "\x1b[31m\x1b[1mhello\x1b[0m", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripANSI(tt.s); got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestValidateChatCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		target  string
		message string
		wantOk  bool
	}{
		{"say valid", "say", "", "Hello!", true},
		{"tell valid", "tell", "player1", "Hi there", true},
		{"tell invalid target", "tell", "p", "Hi", false},
		{"say empty", "say", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateChatCommand(tt.cmd, tt.target, tt.message)
			if ok != tt.wantOk {
				t.Errorf("ValidateChatCommand(%q, %q, %q) = %v, want %v",
					tt.cmd, tt.target, tt.message, ok, tt.wantOk)
			}
		})
	}
}
