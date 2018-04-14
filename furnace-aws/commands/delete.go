package commands

import (
	"log"
	"os"

	"github.com/Skarlso/go-furnace/config"
	awsconfig "github.com/Skarlso/go-furnace/furnace-aws/config"
	"github.com/Skarlso/go-furnace/handle"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
)

// Delete command.
type Delete struct {
}

// Execute defines what this command does.
func (c *Delete) Execute(opts *commander.CommandHelper) {
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	deleteExecute(opts, &client)
}

func deleteExecute(opts *commander.CommandHelper, client *CFClient) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	stackname := awsconfig.Config.Main.Stackname
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
	req := cfClient.Client.DeleteStackRequest(params)
	_, err := req.Send()
	handle.Error(err)
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
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
