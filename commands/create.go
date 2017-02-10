package commands

import (
    "github.com/Skarlso/go-furnace/goaws"
    "github.com/Skarlso/go-furnace/config"
    "github.com/Yitsushi/go-commander"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
    stackConfig := config.LoadCFStackConfig()
    goaws.CreateCF(stackConfig)
}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code to later on.`,
			Arguments:        "",
			Examples:         []string{},
		},
	}
}
