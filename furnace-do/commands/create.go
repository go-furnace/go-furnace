package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/yogsothoth/yogsot"
	yog "github.com/Skarlso/yogsothoth/yogsot"
	commander "github.com/Yitsushi/go-commander"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	yogClient := yog.NewClient()
	template := `
Parameters:
  StackName:
    Description: The name of the stack to deploy
    Type: String
    Default: FurnaceStack
  Port:
    Description: Test port
    Type: Number
    Default: 80

Resources:
  Droplet:
    Name: MyDroplet
    Type: Droplet`
	req := yogsot.CreateStackRequest{
		StackName:    "FurnaceStack",
		TemplateBody: []byte(template),
	}
	res, err := yogClient.CreateStack(req)
	if err != nil {
		config.HandleFatal("error while creating stack:", err)
	}
	log.Println(res)
}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code later on. By default FurnaceStack is used as name.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
