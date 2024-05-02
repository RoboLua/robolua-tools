package main

import (
	"os"
	"robolua-tools/commands"

	"github.com/charmbracelet/log"
)

const (
	release_path = "https://github.com/robolua/robolua/releases/download/v0.1.0/robolua"
)

func main() {
	if len(os.Args) < 2 {
		log.Error("Usage: robolua-tools <command>")
		return
	}

	switch os.Args[1] {
		case "deploy":
			commands.Deploy()
		default:
			log.Error("Unknown command", "command", os.Args[1])
	}
}