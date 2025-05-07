// Package main is the entry point for the MXToolbox clone application.
package main

import (
	"mxclone/cmd/mxclone/commands"
)

func main() {
	// This is just a wrapper that calls the actual CLI implementation
	commands.Execute()
}
