// Package main implements a Matrix-themed Multi-User Dungeon (MUD) game server.
// This file contains terminal/ANSI utilities for text formatting and colorization.
package main

// ANSI escape codes for terminal text formatting and colorization.
// These constants are used throughout the application to provide Matrix-themed
// visual styling with green text, color-coded items by rarity, and clear UI elements.
const (
	Reset = "\033[0m"
	// Basic Colors
	Green   = "\033[32m"
	White   = "\033[97m"
	Gray    = "\033[90m" // Dark Grey for Common
	Red     = "\033[31m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"

	// Rarity Colors
	ColorUncommon = "\033[92m" // Bright Green
	ColorRare     = "\033[96m" // Cyan
	ColorEpic     = "\033[95m" // Magenta/Purple

	Clear = "\033[H\033[2J"
)

// Matrixify wraps text in green ANSI color codes for Matrix-themed output.
// This is the default text style for most game output, creating the characteristic
// green-on-black Matrix aesthetic.
func Matrixify(text string) string {
	return Green + text + Reset
}

// SystemMsg formats a message as a system/operator message with white text.
// System messages are distinguished from regular game text and typically indicate
// important game events or administrative notifications.
func SystemMsg(text string) string {
	return White + "[OPERATOR] " + text + Reset + "\r\n"
}
