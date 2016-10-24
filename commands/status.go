package commands

import (
	"log"

	"github.com/Skarlso/go-aws-mine/goaws"
	"github.com/Yitsushi/go-commander"
)

// EC2Status command.
type EC2Status struct {
}

// Execute defines what this command does.
func (s *EC2Status) Execute(opts *commander.CommandHelper) {
	term := opts.Arg(0)
	if len(term) < 1 {
		log.Fatal("Please provid an EC2 ID.")
	}
	log.Printf("Status for instance id: %s is: %s\n", string(term), goaws.CheckInstanceStatus(string(term)))
}

// NewEC2Status Creates a new CreateEC2 command.
func NewEC2Status(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &EC2Status{},
		Help: &commander.CommandDescriptor{
			Name:             "ec2-status",
			ShortDescription: "Check the status of an EC2 Instance.",
			Arguments:        "<ec2-id>",
			Examples:         []string{"ec2-status id-12345a"},
		},
	}
}
