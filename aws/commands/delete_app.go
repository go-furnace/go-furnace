package commands

import (
	"log"

	awsconfig "github.com/Skarlso/go-furnace/aws/config"
	config "github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/fatih/color"
)

// DeleteApp command.
type DeleteApp struct {
}

// Execute defines what this command does.
func (c *DeleteApp) Execute(opts *commander.CommandHelper) {
	appName := opts.Arg(0)
	if len(appName) < 1 {
		appName = awsconfig.Config.Main.Stackname
	}
	cfg, err := external.LoadDefaultAWSConfig()
	config.CheckError(err)
	cdClient := codedeploy.New(cfg)
	client := CDClient{cdClient}
	deleteApplication(appName, &client)
}

func deleteApplication(appName string, client *CDClient) {
	var yellow = color.New(color.FgYellow).SprintFunc()
	log.Println("Deleting: ", yellow(appName))
	req := client.Client.DeleteApplicationRequest(&codedeploy.DeleteApplicationInput{
		ApplicationName: aws.String(appName),
	})
	_, err := req.Send()
	config.CheckError(err)
}

// NewDeleteApp Creates a new DeleteApp command.
func NewDeleteApp(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &DeleteApp{},
		Help: &commander.CommandDescriptor{
			Name:             "delete-application",
			ShortDescription: "Deletes an Application",
			LongDescription:  `Deletes a CodeDeploy Application complete with DeploymenyGroup and Deploys.`,
			Arguments:        "name",
			Examples:         []string{"delete-application", "delete-application CustomApplicationName"},
		},
	}
}
