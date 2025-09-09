// main.go
package main

import (
	"fmt"
	"log"
	"os"

	"gbe_fork_helper/gbe"
	"gbe_fork_helper/github"
)

// Version of the gbe_fork_helper application
const Version = "v0.1.4"

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// main handles command-line arguments and dispatches commands.
func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	args := os.Args[1:]
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	var err error

	switch command {
	case "apply":
		if len(args) < 3 {
			err = fmt.Errorf("Usage: %s apply <platform> <appid>", os.Args[0])
		} else {
			err = gbe.ApplyGBE(args[1], args[2])
		}
	case "update":
		err = github.UpdateGBE()
	case "version":
		fmt.Println(GetVersion())
	default:
		err = fmt.Errorf("Invalid command: '%s'\n\n", command)
		printUsage()
	}

	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage: gbe_fork_helper <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  apply <platform> <appid> - Apply GBE to Steam API files and configure DLCs")
	fmt.Println("  update                   - Update the GBE fork repository")
	fmt.Println("  version                  - Display the application version")
}
