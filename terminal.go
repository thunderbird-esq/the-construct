package main

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

func Matrixify(text string) string {
	return Green + text + Reset
}

func SystemMsg(text string) string {
	return White + "[OPERATOR] " + text + Reset + "\r\n"
}
