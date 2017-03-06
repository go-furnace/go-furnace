package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"plugin"

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
	p, err := plugin.Open("./plugins/plugins.so")
	if err != nil {
		log.Fatal(err)
	}
	run, err := p.Lookup("RunPlugin")
	if err != nil {
		log.Fatal(err)
	}
	run.(func())()
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Register(commands.NewPush)
	registry.Register(commands.NewDeleteApp)
	registry.Execute()
}
