package display

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	// Define colored printers
	Red     = color.New(color.FgRed).PrintfFunc()
	Green   = color.New(color.FgGreen).PrintfFunc()
	Yellow  = color.New(color.FgYellow).PrintfFunc()
	Blue    = color.New(color.FgBlue).PrintfFunc()
	Magenta = color.New(color.FgMagenta).PrintfFunc()
	Cyan    = color.New(color.FgCyan).PrintfFunc()
	White   = color.New(color.FgWhite).PrintfFunc()

	// Define bold colored printers
	BoldRed     = color.New(color.FgRed, color.Bold).PrintfFunc()
	BoldGreen   = color.New(color.FgGreen, color.Bold).PrintfFunc()
	BoldYellow  = color.New(color.FgYellow, color.Bold).PrintfFunc()
	BoldBlue    = color.New(color.FgBlue, color.Bold).PrintfFunc()
	BoldMagenta = color.New(color.FgMagenta, color.Bold).PrintfFunc()
	BoldCyan    = color.New(color.FgCyan, color.Bold).PrintfFunc()
	Bold        = color.New(color.Bold).PrintfFunc()
)

// Error prints error messages in red
func Error(format string, args ...interface{}) {
	Red(format, args...)
}

// Success prints success messages in green
func Success(format string, args ...interface{}) {
	Green(format, args...)
}

// Warning prints warning messages in yellow
func Warning(format string, args ...interface{}) {
	Yellow(format, args...)
}

// Info prints info messages in cyan
func Info(format string, args ...interface{}) {
	Cyan(format, args...)
}

// Prompt prints a prompt message and returns the user input
func Prompt(promptMsg string) (string, error) {
	Cyan("%s", promptMsg)
	var input string
	_, err := fmt.Scanln(&input)
	return input, err
}

// DisableColors turns off color output - useful for piping to files or logs
func DisableColors() {
	color.NoColor = true
}

// EnableColors turns on color output
func EnableColors() {
	color.NoColor = false
}
