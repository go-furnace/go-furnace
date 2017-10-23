package googlecloud

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
)

// Create commands for google Deployment Manager
type Create struct {
}

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating Deployment Manager.")
	deploymentName := "furnace-stack"
	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	if err != nil {
		log.Fatalf(err.Error())
	}

	d, _ := dm.New(client)
	gConfig := config.LoadGoogleStackConfig()
	log.Println("Config: ", string(gConfig))
	config := dm.ConfigFile{
		Content: string(gConfig),
	}
	targetConfiguration := dm.TargetConfiguration{
		Config: &config,
	}
	deployments := dm.Deployment{
		Name:   deploymentName,
		Target: &targetConfiguration,
	}
	log.Println(deployments)
	ret := d.Deployments.Insert(deploymentName, &deployments)
	log.Println(ret)
	op, err := ret.Do()
	if err != nil {
		log.Fatal("error while doing deployment: ", err)
	}
	for op.Progress != 100 {
		log.Println(op.Progress)
	}
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
