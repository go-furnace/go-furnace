package commands

import (
	awscommands "github.com/Skarlso/go-furnace/commands/aws"
	"github.com/Yitsushi/go-commander"
)

// Aws command.
type Aws struct {
}

// Execute the sub aws commands
func (a *Aws) Execute(opts *commander.CommandHelper) {
	registry := commander.NewCommandRegistry()
	registry.Depth = 1
	registry.Register(awscommands.NewStatus)
	registry.Register(awscommands.NewCreate)
	registry.Register(awscommands.NewDelete)
	registry.Register(awscommands.NewStatus)
	registry.Register(awscommands.NewPush)
	registry.Register(awscommands.NewDeleteApp)
	registry.Register(awscommands.NewUpdate)
	registry.Execute()
}

// NewAws Creates a new Aws Base command
func NewAws(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Aws{},
		Help: &commander.CommandDescriptor{
			Name:             "aws",
			ShortDescription: "Aws based commands",
			LongDescription:  `Main entry point for aws based commands.`,
			Arguments:        "",
			Examples:         []string{"aws create", "aws delete", "aws status", "aws push", "aws update"},
		},
	}
}
