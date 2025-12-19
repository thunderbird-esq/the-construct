package accessibility

import (
	"strings"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s == nil {
		t.Fatal("DefaultSettings returned nil")
	}
	if s.ScreenReaderMode {
		t.Error("ScreenReaderMode should be false by default")
	}
	if s.HighContrast {
		t.Error("HighContrast should be false by default")
	}
	if s.FontScale != 1.0 {
		t.Errorf("FontScale = %f, want 1.0", s.FontScale)
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestGetSettingsDefault(t *testing.T) {
	m := NewManager()
	s := m.GetSettings("nonexistent")
	
	if s == nil {
		t.Fatal("Should return default settings")
	}
	if s.ScreenReaderMode {
		t.Error("Should return default values")
	}
}

func TestSetSettings(t *testing.T) {
	m := NewManager()
	s := &Settings{
		ScreenReaderMode: true,
		HighContrast:     true,
	}
	
	m.SetSettings("player1", s)
	
	got := m.GetSettings("player1")
	if !got.ScreenReaderMode {
		t.Error("ScreenReaderMode should be true")
	}
	if !got.HighContrast {
		t.Error("HighContrast should be true")
	}
}

func TestUpdateSetting(t *testing.T) {
	m := NewManager()
	
	// Update screen reader
	ok := m.UpdateSetting("player1", "screen_reader", true)
	if !ok {
		t.Error("UpdateSetting should succeed")
	}
	
	s := m.GetSettings("player1")
	if !s.ScreenReaderMode {
		t.Error("ScreenReaderMode should be true")
	}
}

func TestUpdateSettingAliases(t *testing.T) {
	m := NewManager()
	
	tests := []struct {
		setting string
		value   interface{}
	}{
		{"screenreader", true},
		{"highcontrast", true},
		{"largetext", true},
		{"reducedmotion", true},
		{"colorblind", "protanopia"},
		{"tts", true},
		{"simplified", true},
		{"fontscale", 1.5},
	}
	
	for _, tt := range tests {
		ok := m.UpdateSetting("player1", tt.setting, tt.value)
		if !ok {
			t.Errorf("UpdateSetting(%s, %v) failed", tt.setting, tt.value)
		}
	}
}

func TestUpdateSettingInvalid(t *testing.T) {
	m := NewManager()
	
	ok := m.UpdateSetting("player1", "nonexistent", true)
	if ok {
		t.Error("Should fail for unknown setting")
	}
}

func TestUpdateSettingFontScaleRange(t *testing.T) {
	m := NewManager()
	
	// Too low
	ok := m.UpdateSetting("player1", "fontscale", 0.1)
	if ok {
		t.Error("Should reject font scale below 0.5")
	}
	
	// Too high
	ok = m.UpdateSetting("player1", "fontscale", 5.0)
	if ok {
		t.Error("Should reject font scale above 3.0")
	}
	
	// Valid
	ok = m.UpdateSetting("player1", "fontscale", 1.5)
	if !ok {
		t.Error("Should accept font scale 1.5")
	}
}

func TestStripANSI(t *testing.T) {
	text := "\033[0;32mGreen text\033[0m"
	result := StripANSI(text)
	
	if strings.Contains(result, "\033") {
		t.Error("Should remove ANSI codes")
	}
	if result != "Green text" {
		t.Errorf("Result = %q, want 'Green text'", result)
	}
}

func TestStripANSIComplex(t *testing.T) {
	text := "\033[1;31m\033[4mBold underlined red\033[0m"
	result := StripANSI(text)
	
	if strings.Contains(result, "\033") {
		t.Error("Should remove all ANSI codes")
	}
}

func TestAddScreenReaderAnnotations(t *testing.T) {
	text := "╔═══╗\n║ Hi ║\n╚═══╝"
	result := AddScreenReaderAnnotations(text)
	
	if strings.Contains(result, "╔") {
		t.Error("Should replace box characters")
	}
	if !strings.Contains(result, "[BOX START]") {
		t.Error("Should add BOX START annotation")
	}
}

func TestAddScreenReaderAnnotationsArrows(t *testing.T) {
	text := "Go → North"
	result := AddScreenReaderAnnotations(text)
	
	if strings.Contains(result, "→") {
		t.Error("Should replace arrows")
	}
	if !strings.Contains(result, "to") {
		t.Error("Should replace → with 'to'")
	}
}

func TestAddScreenReaderAnnotationsExits(t *testing.T) {
	text := "Exits: north, south"
	result := AddScreenReaderAnnotations(text)
	
	if !strings.Contains(result, "[EXITS]") {
		t.Error("Should add EXITS annotation")
	}
}

func TestSimplifyOutput(t *testing.T) {
	text := "Hello\n═══════\n\nWorld"
	result := SimplifyOutput(text)
	
	if strings.Contains(result, "═══") {
		t.Error("Should remove decorative lines")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("Should keep content")
	}
	if !strings.Contains(result, "World") {
		t.Error("Should keep content")
	}
}

func TestIsDecorativeLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"═══════", true},
		{"───────", true},
		{"╔═══╗", true},
		{"Hello", false},
		{"║ Text ║", false},
		{"", true},
		{"   ", true},
	}
	
	for _, tt := range tests {
		result := isDecorativeLine(tt.line)
		if result != tt.expected {
			t.Errorf("isDecorativeLine(%q) = %v, want %v", tt.line, result, tt.expected)
		}
	}
}

func TestProcessText(t *testing.T) {
	settings := &Settings{
		ScreenReaderMode: true,
	}
	
	text := "\033[0;32m╔═══╗\n║ Hi ║\n╚═══╝\033[0m"
	result := ProcessText(text, settings)
	
	if strings.Contains(result, "\033") {
		t.Error("Should strip ANSI in screen reader mode")
	}
	if strings.Contains(result, "╔") {
		t.Error("Should add annotations in screen reader mode")
	}
}

func TestProcessTextNil(t *testing.T) {
	text := "Hello"
	result := ProcessText(text, nil)
	
	if result != text {
		t.Error("Should return unchanged text for nil settings")
	}
}

func TestGetColorScheme(t *testing.T) {
	// Default
	scheme := GetColorScheme(nil)
	if scheme.Name != "Default" {
		t.Errorf("Default scheme name = %s, want Default", scheme.Name)
	}
	
	// High contrast
	settings := &Settings{HighContrast: true}
	scheme = GetColorScheme(settings)
	if scheme.Name != "High Contrast" {
		t.Error("Should return high contrast scheme")
	}
	
	// Colorblind
	settings = &Settings{ColorblindMode: "protanopia"}
	scheme = GetColorScheme(settings)
	if scheme.Name != "Protanopia (Red-Blind)" {
		t.Error("Should return protanopia scheme")
	}
}

func TestColorSchemes(t *testing.T) {
	expectedSchemes := []string{
		"default",
		"high_contrast",
		"protanopia",
		"deuteranopia",
		"tritanopia",
	}
	
	for _, name := range expectedSchemes {
		if _, ok := ColorSchemes[name]; !ok {
			t.Errorf("Missing color scheme: %s", name)
		}
	}
}

func TestApplyColorScheme(t *testing.T) {
	scheme := ColorSchemes["high_contrast"]
	text := "\033[0;32mGreen text\033[0m"
	result := ApplyColorScheme(text, scheme)
	
	// Should replace green with high contrast foreground
	if result == text {
		t.Error("Should modify text with color scheme")
	}
}

func TestApplyColorSchemeNil(t *testing.T) {
	text := "Hello"
	result := ApplyColorScheme(text, nil)
	
	if result != text {
		t.Error("Should return unchanged text for nil scheme")
	}
}

func TestFormatForAccessibility(t *testing.T) {
	settings := &Settings{
		ScreenReaderMode: true,
		SimplifiedOutput: true,
	}
	
	text := "\033[0;32m═══════\nHello World\n═══════\033[0m"
	result := FormatForAccessibility(text, settings)
	
	if strings.Contains(result, "\033") {
		t.Error("Should strip ANSI")
	}
	if strings.Contains(result, "═══") {
		t.Error("Should simplify output")
	}
	if !strings.Contains(result, "Hello World") {
		t.Error("Should keep content")
	}
}

func TestFormatAnnouncementDefault(t *testing.T) {
	a := &Announcement{
		Message:      "You gained XP!",
		ScreenReader: "Experience points increased",
	}
	
	result := FormatAnnouncement(a, nil)
	if result != "You gained XP!" {
		t.Error("Should use Message by default")
	}
}

func TestFormatAnnouncementScreenReader(t *testing.T) {
	a := &Announcement{
		Message:      "You gained XP!",
		ScreenReader: "Experience points increased",
	}
	
	settings := &Settings{ScreenReaderMode: true}
	result := FormatAnnouncement(a, settings)
	if result != "Experience points increased" {
		t.Error("Should use ScreenReader in screen reader mode")
	}
}

func TestGetSettingsList(t *testing.T) {
	settings := &Settings{
		ScreenReaderMode: true,
		HighContrast:     false,
		ColorblindMode:   "protanopia",
		FontScale:        1.5,
	}
	
	result := GetSettingsList(settings)
	
	if !strings.Contains(result, "Screen Reader Mode: ON") {
		t.Error("Should show screen reader as ON")
	}
	if !strings.Contains(result, "High Contrast: OFF") {
		t.Error("Should show high contrast as OFF")
	}
	if !strings.Contains(result, "protanopia") {
		t.Error("Should show colorblind mode")
	}
}

func TestGlobalManager(t *testing.T) {
	if GlobalManager == nil {
		t.Error("GlobalManager should be initialized")
	}
}

func TestProcessOutput(t *testing.T) {
	m := NewManager()
	m.SetSettings("player1", &Settings{ScreenReaderMode: true})
	
	text := "\033[0;32mHello\033[0m"
	result := m.ProcessOutput("player1", text)
	
	if strings.Contains(result, "\033") {
		t.Error("Should process output for player's settings")
	}
}
