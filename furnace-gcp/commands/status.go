package commands

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/handle"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
)

// Status commands for google Deployment Manager
type Status struct {
}

// Execute runs the create command
func (s *Status) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	log.Println("Looking for Deployment under project name: .", keyName(fc.Config.Main.ProjectName))
	status()
}

func status() {
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	handle.Error(err)
	d := NewDeploymentService(ctx, client)
	project := d.Deployments.Get(fc.Config.Main.ProjectName, deploymentName)
	p, err := project.Do()
	if err != nil {
		if err.(*googleapi.Error).Code != 404 {
			handle.Fatal("error while getting deployment: ", err)
		}
		handle.Fatal("Stack not found!", nil)
	}
	if len(p.Manifest) < 1 {
		handle.Fatal("manifest is empty. this usually means that the deployment failed...", errors.New("manifest is empty"))
	}
	manifestID := p.Manifest[strings.LastIndex(p.Manifest, "/")+1 : len(p.Manifest)]
	manifest := d.Manifests.Get(fc.Config.Main.ProjectName, deploymentName, manifestID)
	m, err := manifest.Do()
	handle.Error(err)
	log.Println("Description: ", p.Description)
	log.Println("Name: ", p.Name)
	log.Println("Labels: ", p.Labels)
	log.Println("Selflink: ", p.SelfLink)
	log.Println("Layout: \n", m.Layout)
}

// NewStatus Creates a new status command
func NewStatus(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Status{},
		Help: &commander.CommandDescriptor{
			Name:             "status",
			ShortDescription: "Get the status of an existing Deployment Management group.",
			LongDescription:  `Get the status of an existing Deployment Management group.`,
			Arguments:        "[--config=configFile]",
			Examples:         []string{"status [--config=configFile]"},
		},
	}
}
