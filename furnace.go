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
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(usr.HomeDir, ".config", "go-furnace")); err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("Please create a 'go-furnace' folder under .config.")
		}
	}
}

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewAws)
	registry.Register(commands.NewGoogle)
	registry.Execute()
}
