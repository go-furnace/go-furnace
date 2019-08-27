package commands

import (
	"context"
	"log"
	"os"

	"golang.org/x/oauth2/google"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/handle"
	dm "google.golang.org/api/deploymentmanager/v2"
)

// Update defines and update command struct.
type Update struct {
}

// Execute runs the create command
func (u *Update) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	err := update(fc.Config.Main.ProjectName)
	handle.Error(err)
}

func update(projectName string) error {
	log.Println("Creating Deployment update under project name: .", keyName(projectName))

	deploymentName := fc.Config.Gcp.StackName
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	if err != nil {
		return err
	}
	d := NewDeploymentService(client)
	d.Deployments.Update(projectName, deploymentName, &dm.Deployment{})
	return nil
}

// NewUpdate creates a new update command
func NewUpdate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Update{},
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update updates a Google Deployment",
			LongDescription:  `Using a pre-configured yaml file, update a collection of resources using Deployment Manager Service.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
