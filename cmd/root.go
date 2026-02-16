package cmd

import (
	"fmt"
	"os"
)

var (
	// Version information (set by ldflags)
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type RootOptions struct {
	Config    string
	Verbosity int
}

var rootOpts = RootOptions{
	Verbosity: 1, // info level
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// GetVersion returns the version information
func GetVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}
