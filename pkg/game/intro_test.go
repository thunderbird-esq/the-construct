package game

import (
	"testing"
	"time"
)

func TestDefaultIntroConfig(t *testing.T) {
	config := DefaultIntroConfig()

	if config.Width != 80 {
		t.Errorf("Default width should be 80, got %d", config.Width)
	}
	if config.Height != 24 {
		t.Errorf("Default height should be 24, got %d", config.Height)
	}
	if config.RainFrames <= 0 {
		t.Error("RainFrames should be positive")
	}
	if config.RevealFrames <= 0 {
		t.Error("RevealFrames should be positive")
	}
	if config.FrameDelay <= 0 {
		t.Error("FrameDelay should be positive")
	}
	if config.FinalPause <= 0 {
		t.Error("FinalPause should be positive")
	}
}

func TestBannerDimensions(t *testing.T) {
	// Check SmallBanner fits in 80 columns
	for i, line := range SmallBanner {
		if len(line) > 80 {
			t.Errorf("SmallBanner line %d is %d chars, should be <=80", i, len(line))
		}
	}

	// Check all lines are same length
	if len(SmallBanner) > 0 {
		expectedLen := len(SmallBanner[0])
		for i, line := range SmallBanner {
			if len(line) != expectedLen {
				t.Errorf("SmallBanner line %d is %d chars, expected %d", i, len(line), expectedLen)
			}
		}
	}
}

func TestIntroAnimationGeneratesFrames(t *testing.T) {
	config := IntroConfig{
		Width:        40,
		Height:       12,
		RainFrames:   3,
		RevealFrames: 5,
		FrameDelay:   1 * time.Millisecond,
		FinalPause:   1 * time.Millisecond,
	}

	frames := IntroAnimation(config)

	frameCount := 0
	timeout := time.After(5 * time.Second)

	for {
		select {
		case frame, ok := <-frames:
			if !ok {
				// Channel closed, animation complete
				if frameCount < config.RainFrames+config.RevealFrames {
					t.Errorf("Expected at least %d frames, got %d",
						config.RainFrames+config.RevealFrames, frameCount)
				}
				t.Logf("Animation completed with %d frames", frameCount)
				return
			}
			frameCount++
			if len(frame) == 0 {
				t.Error("Frame should not be empty")
			}
		case <-timeout:
			t.Fatal("Animation timed out")
		}
	}
}

func TestRenderFrameOutput(t *testing.T) {
	screen := make([][]rune, 3)
	brightness := make([][]bool, 3)
	revealed := make([][]bool, 3)

	for i := range screen {
		screen[i] = make([]rune, 5)
		brightness[i] = make([]bool, 5)
		revealed[i] = make([]bool, 5)
		for j := range screen[i] {
			screen[i][j] = 'X'
		}
	}

	// Mark some as bright
	brightness[1][2] = true

	// Mark some as revealed
	revealed[0][0] = true

	output := renderFrame(screen, brightness, revealed)

	// Should contain ANSI codes
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}

	// Should start with home cursor sequence
	if output[0:2] != "\033[" {
		t.Error("Output should start with ANSI escape")
	}

	t.Logf("Frame output length: %d bytes", len(output))
}

func TestRainCharsNotEmpty(t *testing.T) {
	if len(rainChars) == 0 {
		t.Error("rainChars should not be empty")
	}
	t.Logf("Rain character set has %d characters", len(rainChars))
}

func TestANSIConstants(t *testing.T) {
	// Verify ANSI constants are valid escape sequences
	constants := map[string]string{
		"ANSIReset":       ANSIReset,
		"ANSIGreen":       ANSIGreen,
		"ANSIBrightGreen": ANSIBrightGreen,
		"ANSIDimGreen":    ANSIDimGreen,
		"ANSIWhite":       ANSIWhite,
		"ANSIClear":       ANSIClear,
		"ANSIHideCursor":  ANSIHideCursor,
		"ANSIShowCursor":  ANSIShowCursor,
		"ANSIHome":        ANSIHome,
	}

	for name, val := range constants {
		if len(val) == 0 {
			t.Errorf("%s should not be empty", name)
		}
		if val[0] != '\033' {
			t.Errorf("%s should start with ESC (\\033)", name)
		}
	}
}

func TestPlayIntroCallsWriter(t *testing.T) {
	config := IntroConfig{
		Width:        20,
		Height:       5,
		RainFrames:   2,
		RevealFrames: 2,
		FrameDelay:   1 * time.Millisecond,
		FinalPause:   1 * time.Millisecond,
	}

	callCount := 0
	totalBytes := 0

	writer := func(s string) {
		callCount++
		totalBytes += len(s)
	}

	PlayIntro(writer, config)

	if callCount == 0 {
		t.Error("Writer should be called at least once")
	}
	if totalBytes == 0 {
		t.Error("Should write some data")
	}

	t.Logf("PlayIntro called writer %d times, wrote %d total bytes", callCount, totalBytes)
}
