package commands

import (
	"log"

	"github.com/Skarlso/go_aws_mine/aws"
)

// CreateEC2 command.
type CreateEC2 struct {
}

// Execute defines what this command does.
func (c *CreateEC2) Execute() {
	aws.CreateEC2()
	log.Println("Instance created.")
}

// ArgumentDescription describes the arguments for this command.
func (c *CreateEC2) ArgumentDescription() string {
	return ""
}

// Description is the description of this command.
func (c *CreateEC2) Description() string {
	return "Create an EC2 instance to run the server on."
}

// Help displays help information.
func (c *CreateEC2) Help() string {
	return "go_aws_mine create-ec2"
}

// Examples will be displayed by 'help create-ec2'.
func (c *CreateEC2) Examples() []string {
	return []string{"", "test"}
}
