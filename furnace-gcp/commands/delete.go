package commands

import (
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
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
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deleteing Deployment Under Project: ", keyName(fc.Config.Main.ProjectName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	handle.Error(err)
	d2, _ := deploymentmanager.New(client)
	ret := d2.Deployments.Delete(fc.Config.Main.ProjectName, deploymentName)
	_, err = ret.Do()
	if err != nil {
		log.Fatal("error while deleting deployment: ", err)
	}
	waitForDeploymentToFinish(*d2, fc.Config.Main.ProjectName, deploymentName)
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
