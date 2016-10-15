package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go_aws_mine/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func init() {
	// Check if configurations are in the right place. If not, prompt the user to run init.
	log.Println("Checking if go_aws is setup correctly.")
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(usr.HomeDir, ".config", "go_aws_mine")); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Config folder was not found. Please run `./go_aws_mine init` first.")
		}
	}
}

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreateEC2)
	registry.Execute()
}
