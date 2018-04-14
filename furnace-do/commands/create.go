package commands

import (
	"log"
	"os"

	doconfig "github.com/Skarlso/go-furnace/furnace-do/config"
	"github.com/Skarlso/go-furnace/handle"
	"github.com/Skarlso/yogsothoth/yogsot"
	yog "github.com/Skarlso/yogsothoth/yogsot"
	commander "github.com/Yitsushi/go-commander"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := doconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	template := doconfig.LoadDoStackConfig()
	yogClient := yog.NewClient(doconfig.Config.Do.Token)

	req := yogsot.CreateStackRequest{
		StackName:    "FurnaceStack",
		TemplateBody: template,
	}
	res, err := yogClient.CreateStack(req)
	if err != nil {
		handle.Fatal("error while creating stack:", err)
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
