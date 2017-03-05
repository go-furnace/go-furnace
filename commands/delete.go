package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/plugins"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
)

// Delete command.
type Delete struct {
}

// Execute defines what this command does.
func (c *Delete) Execute(opts *commander.CommandHelper) {

	stackname := opts.Arg(0)
	if len(stackname) < 1 {
		stackname = config.STACKNAME
	}

	cyan := color.New(color.FgCyan).SprintFunc()

	log.Printf("Deleting CloudFormation stack with name: %s\n", cyan(stackname))
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	preDeletePlugins := plugins.GetPluginsForEvent(config.PREDELETE)
	log.Println("The following plugins will be triggered pre-delete: ", preDeletePlugins)
	for _, p := range preDeletePlugins {
		p.RunPlugin()
	}
	deleteStack(stackname, &client)
	postDeletePlugins := plugins.GetPluginsForEvent(config.POSTDELETE)
	log.Println("The following plugins will be triggered post-delete: ", postDeletePlugins)
	for _, p := range postDeletePlugins {
		p.RunPlugin()
	}
}

func deleteStack(stackname string, cfClient *CFClient) {
	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackname),
	}
	_, err := cfClient.Client.DeleteStack(params)
	utils.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	utils.WaitForFunctionWithStatusOutput("DELETE_COMPLETE", config.WAITFREQUENCY, func() {
		cfClient.Client.WaitUntilStackDeleteComplete(describeStackInput)
	})
}

// NewDelete Creates a new Delete command.
func NewDelete(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Delete{},
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a stack",
			LongDescription:  `Delete a stack with a given name.`,
			Arguments:        "name",
			Examples:         []string{"delete", "delete MyStackName"},
		},
	}
}
