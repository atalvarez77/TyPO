package logger

import (
	"bytes"
	"strings"
	"testing"
)

// TestLoggerLevels is a white-box test ensuring our level-gating works.
func TestLoggerLevels(t *testing.T) {
	// 1. Setup: Create a buffer to capture output instead of printing to the terminal
	var buf bytes.Buffer

	// 2. Initialize our logger in INFO mode
	l := New(LevelInfo)

	// Override the default output (os.Stdout) with our mock buffer
	l.infoLog.SetOutput(&buf)
	l.debugLog.SetOutput(&buf)

	// 3. Execution: Send an Info and a Debug message
	l.Info("[INFO] This is an info message.")
	l.Debug("[DEBUG] This is a debug message.")

	// 4. Assertion: Read what was actually written
	output := buf.String()

	// We expect the INFO message to be printed
	if !strings.Contains(output, "[INFO] This is an info message.") {
		t.Errorf("Expected output to contain INFO message, got: %s", output)
	}

	// We expect the DEBUG message to be IGNORED because our level is set to LevelInfo
	if strings.Contains(output, "[DEBUG] This is a debug message.") {
		t.Errorf("Did not expect output to contain DEBUG message, got: %s", output)
	}
}
