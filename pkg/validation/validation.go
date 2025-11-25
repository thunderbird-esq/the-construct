// Package validation provides input validation and sanitization utilities.
// This package helps prevent injection attacks and ensures data integrity
// by validating usernames, commands, and sanitizing user input.
package validation

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// validUsername matches alphanumeric usernames with underscores (3-20 characters)
	validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

	// validCommand matches lowercase commands with spaces (1-100 characters)
	validCommand = regexp.MustCompile(`^[a-z ]{1,100}$`)

	// validRoomID matches room identifiers (alphanumeric with underscores and hyphens)
	validRoomID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)
)

// ValidateUsername checks if a username is safe and meets requirements.
// Returns true if the username is 3-20 alphanumeric characters (including underscores).
func ValidateUsername(name string) bool {
	return validUsername.MatchString(name)
}

// ValidateCommand checks if a command string is safe.
// Returns true if the command contains only lowercase letters and spaces.
func ValidateCommand(cmd string) bool {
	return validCommand.MatchString(strings.ToLower(cmd))
}

// ValidateRoomID checks if a room ID is valid.
// Returns true if the room ID contains only safe characters.
func ValidateRoomID(roomID string) bool {
	return validRoomID.MatchString(roomID)
}

// SanitizeInput removes potentially dangerous characters from user input.
// Removes all control characters except newline and tab.
// This prevents terminal escape sequence injection and other attacks.
func SanitizeInput(input string) string {
	// Remove control characters except newline and tab
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1 // Remove character
		}
		return r
	}, input)

	return strings.TrimSpace(cleaned)
}

// SanitizeUsername normalizes a username for storage and comparison.
// Converts to lowercase and trims whitespace.
func SanitizeUsername(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// IsPrintable checks if a string contains only printable characters.
// This can be used for additional validation of user-provided text.
func IsPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}

// ValidatePasswordStrength checks if a password meets minimum security requirements.
// Returns true if password is at least 8 characters long.
// Can be extended to check for complexity requirements.
func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters"
	}

	// Optional: Add complexity checks
	// hasUpper := false
	// hasLower := false
	// hasDigit := false
	// for _, r := range password {
	//     if unicode.IsUpper(r) { hasUpper = true }
	//     if unicode.IsLower(r) { hasLower = true }
	//     if unicode.IsDigit(r) { hasDigit = true }
	// }

	return true, ""
}

// TruncateString safely truncates a string to a maximum length.
// Useful for limiting message lengths and preventing buffer issues.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
