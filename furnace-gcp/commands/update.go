package commands

import (
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/handle"
)

// Update defines and update command struct.
type Update struct {
}

// Execute runs the create command
func (u *Update) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	log.Println("Creating Deployment under project name: .", keyName(fc.Config.Main.ProjectName))
}

// NewUpdate creates a new update command
func NewUpdate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Update{},
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update updates a Google Deployment",
			LongDescription:  `Using a pre-configured yaml file, update a collection of resources using Deployment Manager Service.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
