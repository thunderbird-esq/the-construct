// Package game implements core game mechanics for Matrix MUD.
package game

import (
	"math/rand"
	"strings"
	"time"
)

// Matrix rain characters (katakana-inspired ASCII)
var rainChars = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ0123456789")

// ANSI color codes for intro
const (
	ANSIReset       = "\033[0m"
	ANSIGreen       = "\033[32m"
	ANSIBrightGreen = "\033[92m"
	ANSIDimGreen    = "\033[2;32m"
	ANSIWhite       = "\033[97m"
	ANSIClear       = "\033[H\033[2J"
	ANSIHideCursor  = "\033[?25l"
	ANSIShowCursor  = "\033[?25h"
	ANSIHome        = "\033[H"
)

// IntroConfig holds configuration for the intro animation
type IntroConfig struct {
	Width        int           // Terminal width
	Height       int           // Terminal height
	RainFrames   int           // Number of pure rain frames
	RevealFrames int           // Number of frames for art reveal
	FrameDelay   time.Duration // Delay between frames
	FinalPause   time.Duration // Pause after reveal completes
}

// DefaultIntroConfig returns sensible defaults for telnet clients
func DefaultIntroConfig() IntroConfig {
	return IntroConfig{
		Width:        80,
		Height:       24,
		RainFrames:   30,
		RevealFrames: 60,
		FrameDelay:   50 * time.Millisecond,
		FinalPause:   3 * time.Second,
	}
}

// Banner is the ASCII art banner for THE CONSTRUCT
var Banner = []string{
	"████████╗██╗  ██╗███████╗     ██████╗ ██████╗ ███╗   ██╗███████╗████████╗██████╗ ██╗   ██╗ ██████╗████████╗",
	"╚══██╔══╝██║  ██║██╔════╝    ██╔════╝██╔═══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝",
	"   ██║   ███████║█████╗      ██║     ██║   ██║██╔██╗ ██║███████╗   ██║   ██████╔╝██║   ██║██║        ██║   ",
	"   ██║   ██╔══██║██╔══╝      ██║     ██║   ██║██║╚██╗██║╚════██║   ██║   ██╔══██╗██║   ██║██║        ██║   ",
	"   ██║   ██║  ██║███████╗    ╚██████╗╚██████╔╝██║ ╚████║███████║   ██║   ██║  ██║╚██████╔╝╚██████╗   ██║   ",
	"   ╚═╝   ╚═╝  ╚═╝╚══════╝     ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝  ╚═════╝   ╚═╝   ",
}

// SmallBanner for narrower terminals (80 chars) - pure ASCII
var SmallBanner = []string{
	"+------------------------------------------------------------------------------+",
	"|                                                                              |",
	"|  TTTTT H   H EEEEE     CCC   OOO  N   N  SSS  TTTTT RRRR  U   U  CCC  TTTTT  |",
	"|    T   H   H E        C   C O   O NN  N S       T   R   R U   U C   C   T    |",
	"|    T   HHHHH EEE      C     O   O N N N  SSS    T   RRRR  U   U C       T    |",
	"|    T   H   H E        C   C O   O N  NN     S   T   R  R  U   U C   C   T    |",
	"|    T   H   H EEEEE     CCC   OOO  N   N SSSS    T   R   R  UUU   CCC    T    |",
	"|                                                                              |",
	"+------------------------------------------------------------------------------+",
}

// raindrop represents a single falling character
type raindrop struct {
	x, y   int
	char   rune
	speed  int
	bright bool
}

// IntroAnimation generates and plays the Matrix rain intro animation
// It returns a channel that sends each frame as a string
func IntroAnimation(config IntroConfig) <-chan string {
	frames := make(chan string, 10)

	go func() {
		defer close(frames)

		// Initialize raindrops
		drops := make([]raindrop, config.Width/2)
		for i := range drops {
			drops[i] = raindrop{
				x:      rand.Intn(config.Width),
				y:      rand.Intn(config.Height),
				char:   rainChars[rand.Intn(len(rainChars))],
				speed:  1 + rand.Intn(2),
				bright: rand.Intn(3) == 0,
			}
		}

		// Buffer for screen
		screen := make([][]rune, config.Height)
		brightness := make([][]bool, config.Height)
		revealed := make([][]bool, config.Height)
		for i := range screen {
			screen[i] = make([]rune, config.Width)
			brightness[i] = make([]bool, config.Width)
			revealed[i] = make([]bool, config.Width)
			for j := range screen[i] {
				screen[i][j] = ' '
			}
		}

		// Get banner to reveal
		banner := SmallBanner
		bannerStartY := (config.Height - len(banner)) / 2
		bannerStartX := (config.Width - len(banner[0])) / 2
		if bannerStartX < 0 {
			bannerStartX = 0
		}

		// Phase 1: Pure rain
		for frame := 0; frame < config.RainFrames; frame++ {
			// Clear screen buffer
			for y := range screen {
				for x := range screen[y] {
					screen[y][x] = ' '
					brightness[y][x] = false
				}
			}

			// Update and draw raindrops
			for i := range drops {
				drops[i].y += drops[i].speed
				if drops[i].y >= config.Height {
					drops[i].y = 0
					drops[i].x = rand.Intn(config.Width)
					drops[i].char = rainChars[rand.Intn(len(rainChars))]
					drops[i].bright = rand.Intn(3) == 0
				}

				// Draw raindrop trail
				for t := 0; t < 4; t++ {
					ty := drops[i].y - t
					if ty >= 0 && ty < config.Height && drops[i].x < config.Width {
						if t == 0 {
							screen[ty][drops[i].x] = drops[i].char
							brightness[ty][drops[i].x] = drops[i].bright
						} else {
							screen[ty][drops[i].x] = rainChars[rand.Intn(len(rainChars))]
						}
					}
				}
			}

			frames <- renderFrame(screen, brightness, revealed)
			time.Sleep(config.FrameDelay)
		}

		// Phase 2: Reveal banner through rain
		revealY := 0
		for frame := 0; frame < config.RevealFrames; frame++ {
			// Clear non-revealed areas
			for y := range screen {
				for x := range screen[y] {
					if !revealed[y][x] {
						screen[y][x] = ' '
						brightness[y][x] = false
					}
				}
			}

			// Update raindrops
			for i := range drops {
				drops[i].y += drops[i].speed
				if drops[i].y >= config.Height {
					drops[i].y = 0
					drops[i].x = rand.Intn(config.Width)
					drops[i].char = rainChars[rand.Intn(len(rainChars))]
				}

				// Reveal banner where rain passes
				ty := drops[i].y
				tx := drops[i].x
				if ty >= bannerStartY && ty < bannerStartY+len(banner) {
					bannerY := ty - bannerStartY
					bannerX := tx - bannerStartX
					if bannerX >= 0 && bannerX < len(banner[bannerY]) {
						revealed[ty][tx] = true
					}
				}

				// Draw rain in non-revealed areas
				if !revealed[ty][tx] && ty < config.Height && tx < config.Width {
					screen[ty][tx] = drops[i].char
					brightness[ty][tx] = drops[i].bright
				}
			}

			// Draw revealed banner
			for by, line := range banner {
				y := bannerStartY + by
				if y >= 0 && y < config.Height {
					for bx, ch := range line {
						x := bannerStartX + bx
						if x >= 0 && x < config.Width && revealed[y][x] {
							screen[y][x] = ch
							brightness[y][x] = true
						}
					}
				}
			}

			// Gradually reveal more
			if frame%2 == 0 && revealY < len(banner) {
				y := bannerStartY + revealY
				if y >= 0 && y < config.Height {
					for x := bannerStartX; x < bannerStartX+len(banner[revealY]) && x < config.Width; x++ {
						revealed[y][x] = true
					}
				}
				revealY++
			}

			frames <- renderFrame(screen, brightness, revealed)
			time.Sleep(config.FrameDelay)
		}

		// Final frame with full banner
		for by, line := range banner {
			y := bannerStartY + by
			if y >= 0 && y < config.Height {
				for bx, ch := range line {
					x := bannerStartX + bx
					if x >= 0 && x < config.Width {
						screen[y][x] = ch
						brightness[y][x] = true
						revealed[y][x] = true
					}
				}
			}
		}
		frames <- renderFrame(screen, brightness, revealed)
		time.Sleep(config.FinalPause)
	}()

	return frames
}

// renderFrame converts the screen buffer to an ANSI string
func renderFrame(screen [][]rune, brightness [][]bool, revealed [][]bool) string {
	var sb strings.Builder
	sb.WriteString(ANSIHome)

	for y, row := range screen {
		for x, ch := range row {
			if revealed[y][x] {
				sb.WriteString(ANSIBrightGreen)
			} else if brightness[y][x] {
				sb.WriteString(ANSIBrightGreen)
			} else {
				sb.WriteString(ANSIDimGreen)
			}
			sb.WriteRune(ch)
		}
		sb.WriteString(ANSIReset)
		if y < len(screen)-1 {
			sb.WriteString("\r\n")
		}
	}

	return sb.String()
}

// PlayIntro plays the intro animation to a writer function
// The writer function should write to the client connection
func PlayIntro(writer func(string), config IntroConfig) {
	// Hide cursor and clear screen
	writer(ANSIHideCursor + ANSIClear)

	// Play animation
	for frame := range IntroAnimation(config) {
		writer(frame)
	}

	// Show cursor and prepare for login
	writer(ANSIShowCursor + ANSIClear)
}
