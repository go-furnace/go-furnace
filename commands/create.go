package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	goc "github.com/crewjam/go-cloudformation"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	var cfStack goc.Template
	cfStack = *config.LoadCFStackConfig()
	log.Println("Using configuration: ", cfStack.Parameters["KeyName"].Default)
	log.Println("Stack: ", cfStack)
	// goaws.CreateEC2(ec2Config)
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
