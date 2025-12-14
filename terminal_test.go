package main

import (
	"strings"
	"testing"
)

// TestMatrixify verifies green text wrapping
func TestMatrixify(t *testing.T) {
	result := Matrixify("Hello World")

	if !strings.HasPrefix(result, Green) {
		t.Error("Should start with green color code")
	}
	if !strings.HasSuffix(result, Reset) {
		t.Error("Should end with reset code")
	}
	if !strings.Contains(result, "Hello World") {
		t.Error("Should contain original text")
	}
}

// TestSystemMsg verifies system message formatting
func TestSystemMsg(t *testing.T) {
	result := SystemMsg("Test message")

	if !strings.Contains(result, "[OPERATOR]") {
		t.Error("Should contain OPERATOR prefix")
	}
	if !strings.Contains(result, "Test message") {
		t.Error("Should contain message text")
	}
	if !strings.Contains(result, White) {
		t.Error("Should use white color")
	}
}

// TestApplyThemeInvalid verifies invalid theme handling
func TestApplyThemeInvalid(t *testing.T) {
	input := Green + "Test" + Reset
	result := ApplyTheme(input, "invalid")

	if result != input {
		t.Error("Invalid theme should not modify text")
	}
}

// TestStripColorsAllCodes verifies all color codes are removed
func TestStripColorsAllCodes(t *testing.T) {
	input := Reset + Green + White + Gray + Red + Yellow + Magenta + Cyan +
		ColorUncommon + ColorRare + ColorEpic + "Text"
	result := stripColors(input)

	if result != "Text" {
		t.Errorf("Should be 'Text', got %q", result)
	}
}
