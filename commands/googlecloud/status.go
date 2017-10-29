package googlecloud

import (
	"log"

	config "github.com/Skarlso/go-furnace/config/common"
	fc "github.com/Skarlso/go-furnace/config/google"
	"github.com/Yitsushi/go-commander"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
)

// Status commands for google Deployment Manager
type Status struct {
}

// Execute runs the create command
func (s *Status) Execute(opts *commander.CommandHelper) {
	log.Println("Creating Deployment under project name: .", keyName(fc.GOOGLEPROJECTNAME))
	deploymentName := config.STACKNAME
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	config.CheckError(err)
	d, _ := dm.New(client)
	project := d.Deployments.Get(fc.GOOGLEPROJECTNAME, deploymentName)
	p, err := project.Do()
	config.CheckError(err)
	log.Println(p.Target)
}

// NewStatus Creates a new status command
func NewStatus(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Status{},
		Help: &commander.CommandDescriptor{
			Name:             "status",
			ShortDescription: "Get the status of an existing Deployment Management group.",
			LongDescription:  `I'll write this later`,
			Arguments:        "",
			Examples:         []string{"status"},
		},
	}
}
