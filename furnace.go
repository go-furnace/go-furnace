package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-furnace/commands"
	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/plugins"
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
	// For now, the including of a plugin is done manually.
	prePlug := plugins.MyAwesomePreCreatePlugin{Name: "MyAwesomeSamplePreCreatePlugin"}
	plugins.RegisterPlugin(config.PRECREATE, prePlug)
	postPlug := plugins.MyAwesomePostCreatePlugin{Name: "MyAwesomeSamplePostCreatePlugin"}
	plugins.RegisterPlugin(config.POSTCREATE, postPlug)
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Execute()
}
