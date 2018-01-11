package commands

import (
	"errors"
	"log"
	"strings"

	config "github.com/Skarlso/go-furnace/config"
	fc "github.com/Skarlso/go-furnace/gcp/config"
	"github.com/Yitsushi/go-commander"
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
	if _, ok := opts.Opts["config"]; ok {
		file := opts.Opts["config"]
		if len(file) < 1 {
			config.HandleFatal("Please provide a filename with option -c", errors.New("missing filename"))
		}
		fc.Config.LoadConfiguration(file)
	}
	log.Println("Looking for Deployment under project name: .", keyName(fc.Config.Main.ProjectName))
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	config.CheckError(err)
	d, _ := dm.New(client)
	project := d.Deployments.Get(fc.Config.Main.ProjectName, deploymentName)
	p, err := project.Do()
	if err != nil {
		if err.(*googleapi.Error).Code != 404 {
			config.HandleFatal("error while getting deployment: ", err)
		} else {
			config.HandleFatal("Stack not found!", nil)
		}
	}
	manifestID := p.Manifest[strings.LastIndex(p.Manifest, "/")+1 : len(p.Manifest)]
	manifest := d.Manifests.Get(fc.Config.Main.ProjectName, deploymentName, manifestID)
	m, err := manifest.Do()
	config.CheckError(err)
	log.Println("Description: ", p.Description)
	log.Println("Name: ", p.Name)
	log.Println("Labels: ", p.Labels)
	log.Println("Selflink: ", p.SelfLink)
	log.Println("Layout: \n", m.Layout)
	// Consider getting every resource status?
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
