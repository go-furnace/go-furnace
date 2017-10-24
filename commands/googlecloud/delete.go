package googlecloud

import (
	"log"

	fc "github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
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
	deploymentName := fc.STACKNAME
	log.Println("Deleteing Deployment Under Project: ", keyName(fc.GOOGLEPROJECTNAME))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	if err != nil {
		log.Fatalf(err.Error())
	}
	d2, _ := deploymentmanager.New(client)
	ret := d2.Deployments.Delete(fc.GOOGLEPROJECTNAME, deploymentName)
	_, err = ret.Do()
	if err != nil {
		log.Fatal("error while deleting deployment: ", err)
	}
	utils.WaitForDeploymentToFinish(*d2, deploymentName)
}

// NewDelete Create a new create command
func NewDelete(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Delete{},
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a Google Deployment Manager",
			LongDescription:  `Delete a deployment under a given project id.`,
			Arguments:        "",
			Examples:         []string{"delete"},
		},
	}
}
