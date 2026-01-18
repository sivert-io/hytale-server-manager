package main

import (
	"fmt"
	"os"

	"github.com/sivert-io/hytale-server-manager/src/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		// CLI mode for non-interactive commands
		handleCLI(os.Args[1:])
		return
	}

	// TUI mode
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCLI(args []string) {
	// TODO: Implement CLI commands (start, stop, status, etc.)
	fmt.Println("CLI mode not yet implemented. Use TUI: sudo hsm")
	os.Exit(0)
}
