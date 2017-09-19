package googlecloud

import (
	"github.com/Yitsushi/go-commander"
	"log"
)

// Create commands for google Deployment Manager
type Create struct {
}

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating Deployment Manager.")
}

// NewCreate Creates a new create command
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a Google Deployment Manager",
			LongDescription:  `I'll write this later`,
			Arguments:        "",
			Examples:         []string{"create"},
		},
	}
}
