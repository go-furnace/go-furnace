package commands

import (
	"errors"
	"log"

	config "github.com/Skarlso/go-furnace/config"
	fc "github.com/Skarlso/go-furnace/gcp/config"
	"github.com/Yitsushi/go-commander"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/deploymentmanager/v2"
)

// Delete commands for google Deployment Manager
type Delete struct {
}

// Execute runs the create command
func (d *Delete) Execute(opts *commander.CommandHelper) {
	if _, ok := opts.Opts["config"]; ok {
		file := opts.Opts["config"]
		if len(file) < 1 {
			config.HandleFatal("Please provide a filename with option -c", errors.New("missing filename"))
		}
		fc.Config.LoadConfiguration(file)
	}
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deleteing Deployment Under Project: ", keyName(fc.Config.Main.ProjectName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	config.CheckError(err)
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
			Arguments:        "[--config=configFile]",
			Examples:         []string{"delete [--config=configFile]"},
		},
	}
}
