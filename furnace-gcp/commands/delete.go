package commands

import (
	"errors"
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/furnace-gcp/plugins"
	"github.com/go-furnace/go-furnace/handle"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/deploymentmanager/v2"
)

// Delete commands for google Deployment Manager
type Delete struct {
}

// Execute runs the create command
func (d *Delete) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	handle.Error(err)
	ds := NewDeploymentService(client)
	delete(ds)
	log.Println("Deleteing Deployment Under Project: ", keyName(fc.Config.Main.ProjectName))
}

func delete(d DeploymentmanagerService) error {
	ret := d.Deployments.Delete(fc.Config.Main.ProjectName, fc.Config.Gcp.StackName)
	if ret == nil {
		return errors.New("return of delete was nil")
	}
	plugins.RunPreDeletePlugins(fc.Config.Gcp.StackName)
	_, err := ret.Do()
	if err != nil {
		return err
	}
	waitForDeploymentToFinish(*d.Service, fc.Config.Main.ProjectName, fc.Config.Gcp.StackName)
	plugins.RunPostDeletePlugins(fc.Config.Gcp.StackName)
	return nil
}

// NewDelete Create a new create command
func NewDelete(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Delete{},
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a Google Deployment Manager",
			LongDescription:  `Delete a deployment under a given project id.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
