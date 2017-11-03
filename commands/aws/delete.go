package awscommands

import (
	"log"

	awsconfig "github.com/Skarlso/go-furnace/config/aws"
	config "github.com/Skarlso/go-furnace/config/common"
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
	sess := session.New(&aws.Config{Region: aws.String(awsconfig.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	deleteExecute(opts, &client)
}

func deleteExecute(opts *commander.CommandHelper, client *CFClient) {
	stackname := config.STACKNAME
	cyan := color.New(color.FgCyan).SprintFunc()
	log.Printf("Deleting CloudFormation stack with name: %s\n", cyan(stackname))
	for _, p := range awsconfig.PluginRegistry[awsconfig.PREDELETE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
	deleteStack(stackname, client)
	for _, p := range awsconfig.PluginRegistry[awsconfig.POSTDELETE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
}

func deleteStack(stackname string, cfClient *CFClient) {
	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackname),
	}
	_, err := cfClient.Client.DeleteStack(params)
	config.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("DELETE_COMPLETE", config.WAITFREQUENCY, func() {
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
			Arguments:        "",
			Examples:         []string{""},
		},
	}
}
