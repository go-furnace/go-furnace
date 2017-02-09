package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/goaws"
	"github.com/Yitsushi/go-commander"
)

// TerminateEC2 command.
type TerminateEC2 struct {
}

// Execute defines what this command does.
func (c *TerminateEC2) Execute(opts *commander.CommandHelper) {
	term := opts.Arg(0)
	if len(term) < 1 {
		log.Fatal("Please provid an EC2 ID.")
	}
	goaws.TerminateEC2(string(term))
}

// NewTerminateEC2 Terminates a new TerminateEC2 command.
func NewTerminateEC2(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &TerminateEC2{},
		Help: &commander.CommandDescriptor{
			Name:             "terminate-ec2",
			ShortDescription: "Terminate an EC2 instance.",
			Arguments:        "<ec2-id>",
			Examples:         []string{"terminate-ec2 i-09491asdf"},
		},
	}
}
