package awscommands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/fatih/color"
)

// DeleteApp command.
type DeleteApp struct {
}

// Execute defines what this command does.
func (c *DeleteApp) Execute(opts *commander.CommandHelper) {
	appName := opts.Arg(0)
	if len(appName) < 1 {
		appName = config.STACKNAME
	}
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cdClient := codedeploy.New(sess, nil)
	client := CDClient{cdClient}
	deleteApplication(appName, &client)
}

func deleteApplication(appName string, client *CDClient) {
	var yellow = color.New(color.FgYellow).SprintFunc()
	log.Println("Deleting: ", yellow(appName))
	_, err := client.Client.DeleteApplication(&codedeploy.DeleteApplicationInput{
		ApplicationName: aws.String(appName),
	})
	utils.CheckError(err)
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
