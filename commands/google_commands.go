package commands

import (
	googlecommands "github.com/Skarlso/go-furnace/commands/googlecloud"
	"github.com/Yitsushi/go-commander"
)

// Google command.
type Google struct {
}

// Execute the sub aws commands
func (a *Google) Execute(opts *commander.CommandHelper) {
	registry := commander.NewCommandRegistry()
	registry.Register(googlecommands.NewCreate)
	registry.Depth = 1
	registry.Execute()
}

// NewGoogle Creates a new Aws Base command
func NewGoogle(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Google{},
		Help: &commander.CommandDescriptor{
			Name:             "google",
			ShortDescription: "Google based commands",
			LongDescription:  `Main entry point for google based commands.`,
			Arguments:        "",
			Examples:         []string{"google create", "google delete", "google status"},
		},
	}
}
