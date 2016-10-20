package commands

import (
	"log"

	"github.com/Skarlso/go_aws_mine/config"
	"github.com/Yitsushi/go-commander"
)

// CreateEC2 command.
type CreateEC2 struct {
}

// Execute defines what this command does.
func (c *CreateEC2) Execute(opts *commander.CommandHelper) {
	ec2Config := config.CreateEC2Config()
	log.Println(ec2Config.DryRun)
}

// NewCreateEC2 Creates a new CreateEC2 command.
func NewCreateEC2(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &CreateEC2{},
		Help: &commander.CommandDescriptor{
			Name:             "create-ec2",
			ShortDescription: "Create an EC2 instance.",
			LongDescription: `Allocate a t2.large ( or whatever is configured ) instance
on which a minecraft server will be running.`,
			Arguments: "",
			Examples:  []string{},
		},
	}
}
