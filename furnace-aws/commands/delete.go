package commands

import (
	"context"
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/furnace-aws/plugins"
	"github.com/go-furnace/go-furnace/handle"
)

// Delete command.
type Delete struct {
	client *CFClient
}

// Execute defines what this command does.
func (d *Delete) Execute(opts *commander.CommandHelper) {
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
	deleteStack(stackname, d.client)
	plugins.RunPostDeletePlugins(stackname)
}

func deleteStack(stackname string, cfClient *CFClient) {
	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackname),
	}
	ctx := context.Background()
	req := cfClient.Client.DeleteStackRequest(params)
	plugins.RunPreDeletePlugins(stackname)
	_, err := req.Send(ctx)
	handle.Error(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("DELETE_COMPLETE", config.WAITFREQUENCY, func() {
		err := cfClient.Client.WaitUntilStackDeleteComplete(ctx, describeStackInput)
		if err != nil {
			return
		}
	})
}

// NewDelete Creates a new Delete command.
func NewDelete(appName string) *commander.CommandWrapper {
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	d := Delete{&client}
	return &commander.CommandWrapper{
		Handler: &d,
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a stack",
			LongDescription:  `Delete a stack with a given name.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
