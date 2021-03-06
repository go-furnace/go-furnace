package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws/external"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/furnace-aws/plugins"
	"github.com/go-furnace/go-furnace/handle"
)

// Create command.
type Create struct {
	client *CFClient
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	stackname := awsconfig.Config.Main.Stackname
	template := awsconfig.LoadCFStackConfig()
	stacks := create(stackname, template, c.client)
	plugins.RunPostCreatePlugins(stackname)
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(stacks[0].StackStatus))
	} else {
		handle.Fatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

// create will create a full stack and encapsulate the functionality of
// the create command.
func create(stackname string, template []byte, cfClient *CFClient) []cloudformation.Stack {
	validResp := cfClient.validateTemplate(template)
	if validResp == nil {
		log.Println("The response from AWS to validate was nil.")
		return nil
	}
	stackParameters := gatherParameters(os.Stdin, validResp.ValidateTemplateOutput)
	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackname),
		Capabilities: []cloudformation.Capability{cloudformation.CapabilityCapabilityIam},
		Parameters:   stackParameters,
		TemplateBody: aws.String(string(template)),
	}
	plugins.RunPreCreatePlugins(stackname)
	resp := cfClient.createStack(stackInputParams)
	if resp == nil {
		log.Println("The response to create stack from AWS was nil.")
		return nil
	}
	log.Println("Create stack response: ", resp)
	cfClient.waitForStackCreateCompleteStatus(context.Background(), stackname)
	descResp := cfClient.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	if descResp != nil {
		return descResp.Stacks
	}
	return nil
}

func (cf *CFClient) waitForStackCreateCompleteStatus(ctx context.Context, stackname string) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("CREATE_COMPLETE", config.WAITFREQUENCY, func() {
		err := cf.Client.WaitUntilStackCreateComplete(ctx, describeStackInput)
		if err != nil {
			return
		}
	})
}

func (cf *CFClient) createStack(stackInputParams *cloudformation.CreateStackInput) *cloudformation.CreateStackResponse {
	log.Println("Creating Stack with name: ", keyName(*stackInputParams.StackName))
	req := cf.Client.CreateStackRequest(stackInputParams)
	resp, err := req.Send(context.Background())
	handle.Error(err)
	return resp
}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	log.Println("Creating cloud formation session.")
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	c := Create{}
	c.client = &CFClient{cfClient}
	return &commander.CommandWrapper{
		Handler: &c,
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code later on. By default FurnaceStack is used as name.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
