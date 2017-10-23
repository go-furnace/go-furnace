package googlecloud

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
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
	log.Println("Deleteing Deployment Manager.")
	deploymentName := "furnace-stack"
	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, deploymentmanager.NdevCloudmanScope)
	if err != nil {
		log.Fatalf(err.Error())
	}

	d2, _ := deploymentmanager.New(client)
	gConfig := config.LoadGoogleStackConfig()
	log.Println("Config: ", string(gConfig))
	ret := d2.Deployments.Delete("<PROJECT_ID>", deploymentName)
	op, err := ret.Do()
	if err != nil {
		log.Fatal("error while deleting deployment: ", err)
	}
	log.Println(op)
}

// NewDelete Create a new create command
func NewDelete(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Delete{},
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a Google Deployment Manager",
			LongDescription:  `I'll write this later`,
			Arguments:        "",
			Examples:         []string{"delete"},
		},
	}
}
