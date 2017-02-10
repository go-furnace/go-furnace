package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-furnace/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func init() {
	// Check if configurations are in the right place. If not, prompt the user to run init.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Create the environment first if it doesn't exists instead of having it a command.
	if _, err := os.Stat(filepath.Join(usr.HomeDir, ".config", "go-furnace")); err != nil {
		if os.IsNotExist(err) {
			i := commands.Init{}
			i.Execute(nil)
		}
	}
}

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Execute()
}
