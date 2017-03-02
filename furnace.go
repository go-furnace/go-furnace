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
	preCreatePlugins := []plugins.Plugin{
		plugins.MyAwesomePreCreatePlugin{Name: "SamplePreCreatePlugin"},
	}
	postCreatePlugins := []plugins.Plugin{
		plugins.MyAwesomePostCreatePlugin{Name: "SamplePostCreatePlugin"},
	}
	plugins.RegisterPlugin(config.PRECREATE, preCreatePlugins)
	plugins.RegisterPlugin(config.POSTCREATE, postCreatePlugins)
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Register(commands.NewPush)
	registry.Execute()
}
