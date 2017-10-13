package googlecloud

import (
	"log"

	"github.com/Yitsushi/go-commander"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	dm "google.golang.org/api/deploymentmanager/v2"
)

// Create commands for google Deployment Manager
type Create struct {
}

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating Deployment Manager.")
	// Use oauth2.NoContext if there isn't a good context to pass in.
	ctx := context.TODO()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println(client)
	d, _ := dm.New(client)
	log.Println(d)
}

// NewCreate Creates a new create command
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a Google Deployment Manager",
			LongDescription:  `I'll write this later`,
			Arguments:        "",
			Examples:         []string{"create"},
		},
	}
}
