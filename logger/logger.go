package logger

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"os"
)

var (
	// Red color red
	Red = color.Red
	// Purple color purple
	Purple = color.Magenta
	// Green color green
	Green = color.LightGreen
	// Yellow color yellow
	Yellow = color.Yellow
)

// Logf logs to console with tag and content
func Logf(tag string, content string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, fmt.Sprintf("[%s] - ", Purple.Sprint(tag))+content+"\n", a...)
}

// Warn logs a warning to console
func Warn(err interface{}) {
	// Handle strings being passed by creating an error type
	if err != nil {
		if e, ok := err.(string); ok {
			err = errors.New(e)
		}
		fmt.Printf("[%s] - %v\n", Yellow.Sprint("WARN"), err)
	}
}

// Die handles errors and exits
func Die(err interface{}) {
	// Handle strings being passed by creating an error type
	if err != nil {
		if e, ok := err.(string); ok {
			err = errors.New(e)
		}
		fmt.Printf("[%s] - %v\n", Red.Sprint("ERR"), err)
		os.Exit(1)
	}
}
