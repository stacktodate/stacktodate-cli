package helpers

import (
	"fmt"
	"os"
)

// ExitWithError prints formatted error message to stderr and exits with given code
func ExitWithError(exitCode int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(exitCode)
}

// ExitOnError is a convenience wrapper that exits with code 1
func ExitOnError(err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		ExitWithError(1, format+": %v", args...)
	}
}
