// Package accessibility provides accessibility features for Matrix MUD.
// Includes screen reader support, high contrast themes, and reduced motion options.
package accessibility

import (
	"regexp"
	"strings"
	"sync"
)

// Settings represents accessibility preferences for a player
type Settings struct {
	ScreenReaderMode  bool    `json:"screen_reader_mode"`
	HighContrast      bool    `json:"high_contrast"`
	LargeText         bool    `json:"large_text"`
	ReducedMotion     bool    `json:"reduced_motion"`
	ColorblindMode    string  `json:"colorblind_mode"` // none, protanopia, deuteranopia, tritanopia
	TextToSpeech      bool    `json:"text_to_speech"`
	SimplifiedOutput  bool    `json:"simplified_output"`
	DisableAnimations bool    `json:"disable_animations"`
	FontScale         float64 `json:"font_scale"` // 1.0 = normal, 1.5 = 150%, etc.
}

// DefaultSettings returns default accessibility settings
func DefaultSettings() *Settings {
	return &Settings{
		ScreenReaderMode:  false,
		HighContrast:      false,
		LargeText:         false,
		ReducedMotion:     false,
		ColorblindMode:    "none",
		TextToSpeech:      false,
		SimplifiedOutput:  false,
		DisableAnimations: false,
		FontScale:         1.0,
	}
}

// Manager manages accessibility settings for all players
type Manager struct {
	mu       sync.RWMutex
	settings map[string]*Settings // player name -> settings
}

// NewManager creates a new accessibility manager
func NewManager() *Manager {
	return &Manager{
		settings: make(map[string]*Settings),
	}
}

// GetSettings returns settings for a player
func (m *Manager) GetSettings(playerName string) *Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.settings[playerName]; ok {
		return s
	}
	return DefaultSettings()
}

// SetSettings sets settings for a player
func (m *Manager) SetSettings(playerName string, settings *Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings[playerName] = settings
}

// UpdateSetting updates a single setting
func (m *Manager) UpdateSetting(playerName, setting string, value interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.settings[playerName]
	if !ok {
		s = DefaultSettings()
		m.settings[playerName] = s
	}

	switch setting {
	case "screen_reader", "screenreader":
		if v, ok := value.(bool); ok {
			s.ScreenReaderMode = v
			return true
		}
	case "high_contrast", "highcontrast":
		if v, ok := value.(bool); ok {
			s.HighContrast = v
			return true
		}
	case "large_text", "largetext":
		if v, ok := value.(bool); ok {
			s.LargeText = v
			return true
		}
	case "reduced_motion", "reducedmotion":
		if v, ok := value.(bool); ok {
			s.ReducedMotion = v
			return true
		}
	case "colorblind":
		if v, ok := value.(string); ok {
			s.ColorblindMode = v
			return true
		}
	case "tts", "text_to_speech":
		if v, ok := value.(bool); ok {
			s.TextToSpeech = v
			return true
		}
	case "simplified":
		if v, ok := value.(bool); ok {
			s.SimplifiedOutput = v
			return true
		}
	case "animations":
		if v, ok := value.(bool); ok {
			s.DisableAnimations = !v // "animations off" = DisableAnimations true
			return true
		}
	case "font_scale", "fontscale":
		if v, ok := value.(float64); ok && v >= 0.5 && v <= 3.0 {
			s.FontScale = v
			return true
		}
	}

	return false
}

// ANSI escape code regex
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ProcessOutput processes output text based on accessibility settings
func (m *Manager) ProcessOutput(playerName, text string) string {
	settings := m.GetSettings(playerName)
	return ProcessText(text, settings)
}

// ProcessText processes text with given settings
func ProcessText(text string, settings *Settings) string {
	if settings == nil {
		return text
	}

	// Screen reader mode: strip ANSI, add structure
	if settings.ScreenReaderMode {
		text = StripANSI(text)
		text = AddScreenReaderAnnotations(text)
	}

	// Simplified output: reduce verbosity
	if settings.SimplifiedOutput {
		text = SimplifyOutput(text)
	}

	return text
}

// StripANSI removes all ANSI escape codes from text
func StripANSI(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

// AddScreenReaderAnnotations adds semantic annotations for screen readers
func AddScreenReaderAnnotations(text string) string {
	// Add section markers for common patterns
	text = strings.ReplaceAll(text, "╔", "[BOX START]")
	text = strings.ReplaceAll(text, "╚", "[BOX END]")
	text = strings.ReplaceAll(text, "║", "|")
	text = strings.ReplaceAll(text, "═", "-")
	text = strings.ReplaceAll(text, "╠", "|")
	text = strings.ReplaceAll(text, "╣", "|")
	text = strings.ReplaceAll(text, "╗", "")
	text = strings.ReplaceAll(text, "╝", "")

	// Replace common symbols
	text = strings.ReplaceAll(text, "→", "to")
	text = strings.ReplaceAll(text, "←", "from")
	text = strings.ReplaceAll(text, "↑", "up")
	text = strings.ReplaceAll(text, "↓", "down")
	text = strings.ReplaceAll(text, "•", "*")
	text = strings.ReplaceAll(text, "█", "#")
	text = strings.ReplaceAll(text, "▓", "#")
	text = strings.ReplaceAll(text, "░", ".")

	// Add descriptive headers
	if strings.Contains(text, "HP:") {
		text = "[STATUS] " + text
	}
	if strings.Contains(text, "Exits:") {
		text = strings.Replace(text, "Exits:", "[EXITS]", 1)
	}

	return text
}

// SimplifyOutput reduces verbose output
func SimplifyOutput(text string) string {
	// Remove decorative lines
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		// Skip purely decorative lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if isDecorativeLine(trimmed) {
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// isDecorativeLine checks if a line is purely decorative
func isDecorativeLine(line string) bool {
	// Lines that are all box-drawing or decorative characters
	decorativeChars := "═─╔╗╚╝║╠╣╬░▒▓█▀▄"
	for _, r := range line {
		if !strings.ContainsRune(decorativeChars, r) && r != ' ' {
			return false
		}
	}
	return true
}

// ColorScheme represents a color scheme for accessibility
type ColorScheme struct {
	Name        string
	Foreground  string
	Background  string
	Highlight   string
	Error       string
	Success     string
	Warning     string
	Info        string
	Muted       string
}

// Color schemes for different modes
var ColorSchemes = map[string]*ColorScheme{
	"default": {
		Name:       "Default",
		Foreground: "\033[0;32m",  // Green
		Background: "",
		Highlight:  "\033[1;32m",  // Bright Green
		Error:      "\033[0;31m",  // Red
		Success:    "\033[0;32m",  // Green
		Warning:    "\033[0;33m",  // Yellow
		Info:       "\033[0;36m",  // Cyan
		Muted:      "\033[0;90m",  // Gray
	},
	"high_contrast": {
		Name:       "High Contrast",
		Foreground: "\033[1;37m",  // Bright White
		Background: "",
		Highlight:  "\033[1;37m",  // Bright White
		Error:      "\033[1;31m",  // Bright Red
		Success:    "\033[1;32m",  // Bright Green
		Warning:    "\033[1;33m",  // Bright Yellow
		Info:       "\033[1;36m",  // Bright Cyan
		Muted:      "\033[0;37m",  // White
	},
	"protanopia": {
		Name:       "Protanopia (Red-Blind)",
		Foreground: "\033[0;36m",  // Cyan
		Background: "",
		Highlight:  "\033[1;36m",  // Bright Cyan
		Error:      "\033[0;35m",  // Magenta
		Success:    "\033[0;34m",  // Blue
		Warning:    "\033[0;33m",  // Yellow
		Info:       "\033[0;37m",  // White
		Muted:      "\033[0;90m",  // Gray
	},
	"deuteranopia": {
		Name:       "Deuteranopia (Green-Blind)",
		Foreground: "\033[0;34m",  // Blue
		Background: "",
		Highlight:  "\033[1;34m",  // Bright Blue
		Error:      "\033[0;35m",  // Magenta
		Success:    "\033[0;36m",  // Cyan
		Warning:    "\033[0;33m",  // Yellow
		Info:       "\033[0;37m",  // White
		Muted:      "\033[0;90m",  // Gray
	},
	"tritanopia": {
		Name:       "Tritanopia (Blue-Blind)",
		Foreground: "\033[0;32m",  // Green
		Background: "",
		Highlight:  "\033[1;32m",  // Bright Green
		Error:      "\033[0;31m",  // Red
		Success:    "\033[0;32m",  // Green
		Warning:    "\033[0;31m",  // Red (instead of yellow)
		Info:       "\033[0;35m",  // Magenta
		Muted:      "\033[0;90m",  // Gray
	},
}

// GetColorScheme returns the appropriate color scheme for settings
func GetColorScheme(settings *Settings) *ColorScheme {
	if settings == nil {
		return ColorSchemes["default"]
	}

	if settings.HighContrast {
		return ColorSchemes["high_contrast"]
	}

	if settings.ColorblindMode != "" && settings.ColorblindMode != "none" {
		if scheme, ok := ColorSchemes[settings.ColorblindMode]; ok {
			return scheme
		}
	}

	return ColorSchemes["default"]
}

// ApplyColorScheme applies a color scheme to text
func ApplyColorScheme(text string, scheme *ColorScheme) string {
	if scheme == nil {
		return text
	}

	// Replace standard colors with scheme colors
	text = strings.ReplaceAll(text, "\033[0;32m", scheme.Foreground)
	text = strings.ReplaceAll(text, "\033[1;32m", scheme.Highlight)
	text = strings.ReplaceAll(text, "\033[0;31m", scheme.Error)
	text = strings.ReplaceAll(text, "\033[0;33m", scheme.Warning)
	text = strings.ReplaceAll(text, "\033[0;36m", scheme.Info)

	return text
}

// FormatForAccessibility formats text with all accessibility transformations
func FormatForAccessibility(text string, settings *Settings) string {
	if settings == nil {
		return text
	}

	// Process text first
	text = ProcessText(text, settings)

	// Apply color scheme if not in screen reader mode
	if !settings.ScreenReaderMode {
		scheme := GetColorScheme(settings)
		text = ApplyColorScheme(text, scheme)
	}

	return text
}

// Announcement creates an accessible announcement
type Announcement struct {
	Priority    string // low, normal, high, critical
	Category    string // combat, system, social, etc.
	Message     string
	ScreenReader string // Optional screen reader-friendly version
}

// FormatAnnouncement formats an announcement for accessibility
func FormatAnnouncement(a *Announcement, settings *Settings) string {
	if settings != nil && settings.ScreenReaderMode && a.ScreenReader != "" {
		return a.ScreenReader
	}
	return a.Message
}

// GetSettingsList returns a formatted list of all settings
func GetSettingsList(settings *Settings) string {
	var sb strings.Builder
	sb.WriteString("Accessibility Settings:\n")
	sb.WriteString("───────────────────────\n")
	sb.WriteString(formatBoolSetting("Screen Reader Mode", settings.ScreenReaderMode))
	sb.WriteString(formatBoolSetting("High Contrast", settings.HighContrast))
	sb.WriteString(formatBoolSetting("Large Text", settings.LargeText))
	sb.WriteString(formatBoolSetting("Reduced Motion", settings.ReducedMotion))
	sb.WriteString(formatBoolSetting("Text to Speech", settings.TextToSpeech))
	sb.WriteString(formatBoolSetting("Simplified Output", settings.SimplifiedOutput))
	sb.WriteString(formatBoolSetting("Disable Animations", settings.DisableAnimations))
	sb.WriteString("Colorblind Mode: " + settings.ColorblindMode + "\n")
	sb.WriteString("Font Scale: " + formatFloat(settings.FontScale) + "x\n")
	return sb.String()
}

func formatBoolSetting(name string, value bool) string {
	status := "OFF"
	if value {
		status = "ON"
	}
	return name + ": " + status + "\n"
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return strings.TrimSuffix(strings.TrimSuffix(
			strings.Replace(
				strings.Replace(
					string(rune('0'+int(f))),
					"1", "1.0", 1),
				"2", "2.0", 1),
			".0"), ".0")
	}
	// Simple float formatting
	intPart := int(f)
	decPart := int((f - float64(intPart)) * 10)
	return string(rune('0'+intPart)) + "." + string(rune('0'+decPart))
}

// Global manager instance
var GlobalManager = NewManager()
