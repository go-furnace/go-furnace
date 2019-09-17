package commands

import (
	"errors"
	"fmt"
	"log"
	"net/http"
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
	client *http.Client
	ctx    context.Context
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
	ds := NewDeploymentService(d.ctx, d.client)
	err := deleteDeployment(ds)
	handle.Error(err)
	log.Println("Deleting Deployment Under Project: ", keyName(fc.Config.Main.ProjectName))
}

func deleteDeployment(d DeploymentmanagerService) error {
	ret := d.Deployments.Delete(fc.Config.Main.ProjectName, fc.Config.Gcp.StackName)
	if ret == nil {
		return errors.New("return of deleteDeployment was nil")
	}
	plugins.RunPreDeletePlugins(fc.Config.Gcp.StackName)
	_, err := ret.Do()
	if err != nil {
		return err
	}
	fmt.Printf("d: %#v\n", d)
	waitForDeploymentToFinish(d, fc.Config.Main.ProjectName, fc.Config.Gcp.StackName)
	plugins.RunPostDeletePlugins(fc.Config.Gcp.StackName)
	return nil
}

// NewDelete Create a new create command
func NewDelete(appName string) *commander.CommandWrapper {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	handle.Error(err)
	d := Delete{client: client, ctx: ctx}
	return &commander.CommandWrapper{
		Handler: &d,
		Help: &commander.CommandDescriptor{
			Name:             "deleteDeployment",
			ShortDescription: "Delete a Google Deployment Manager",
			LongDescription:  `Delete a deployment under a given project id.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
