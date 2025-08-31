// main.go
package main

import (
	"fmt"
	"log"
	"os"

	"gbe_fork_helper/gbe"
	"gbe_fork_helper/github"
	"gbe_fork_helper/steam"
)

// main handles command-line arguments and dispatches commands.
func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: gbe_fork_helper <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  apply <platform> - Apply GBE to Steam API files")
		fmt.Println("  update           - Update the GBE fork repository")
		fmt.Println("  dlc <appid>      - Fetch DLCs for a given AppID")
		os.Exit(1)
	}

	command := args[0]
	var err error

	switch command {
	case "apply":
		if len(args) < 2 {
			err = fmt.Errorf("Usage: %s apply <platform>", os.Args[0])
		} else {
			err = gbe.ApplyGBE(args[1])
		}
	case "update":
		err = github.UpdateGBE()
	case "dlc":
		if len(args) < 2 {
			err = fmt.Errorf("Usage: %s dlc <appid>", os.Args[0])
		} else {
			err = steam.FetchDLCs(args[1])
		}
	default:
		err = fmt.Errorf("Invalid command: '%s'", command)
	}

	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
