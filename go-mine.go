package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-aws-mine/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func init() {
	// Check if configurations are in the right place. If not, prompt the user to run init.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Create the environment first if it doesn't exists instead of having it a command.
	if _, err := os.Stat(filepath.Join(usr.HomeDir, ".config", "go-aws-mine")); err != nil {
		if os.IsNotExist(err) {
			// TODO: Run init here.
			i := commands.Init{}
			i.Execute(nil)
		}
	}
}

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreateEC2)
	registry.Register(commands.NewEC2Status)
	registry.Register(commands.NewTerminateEC2)
	registry.Execute()
}
