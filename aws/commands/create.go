package commands

import (
	"fmt"
	"log"
	"os"

	awsconfig "github.com/Skarlso/go-furnace/aws/config"
	config "github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	cfg, err := external.LoadDefaultAWSConfig()
	config.CheckError(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	createExecute(opts, &client)
}

func createExecute(opts *commander.CommandHelper, client *CFClient) {
	stackname := config.STACKNAME
	template := awsconfig.LoadCFStackConfig()
	for _, p := range awsconfig.PluginRegistry[awsconfig.PRECREATE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
	stacks := create(stackname, template, client)
	for _, p := range awsconfig.PluginRegistry[awsconfig.POSTCREATE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(stacks[0].StackStatus))
	} else {
		config.HandleFatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

// create will create a full stack and encapsulate the functionality of
// the create command.
func create(stackname string, template []byte, cfClient *CFClient) []cloudformation.Stack {
	validResp := cfClient.validateTemplate(template)
	stackParameters := gatherParameters(os.Stdin, validResp)
	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackname),
		Capabilities: []cloudformation.Capability{cloudformation.CapabilityCapabilityIam},
		Parameters:   stackParameters,
		TemplateBody: aws.String(string(template)),
	}
	resp := cfClient.createStack(stackInputParams)
	log.Println("Create stack response: ", resp)
	cfClient.waitForStackCreateCompleteStatus(stackname)
	descResp := cfClient.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	if descResp != nil {
		return descResp.Stacks
	}
	return nil
}

func (cf *CFClient) waitForStackCreateCompleteStatus(stackname string) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("CREATE_COMPLETE", config.WAITFREQUENCY, func() {
		cf.Client.WaitUntilStackCreateComplete(describeStackInput)
	})
}

func (cf *CFClient) createStack(stackInputParams *cloudformation.CreateStackInput) *cloudformation.CreateStackOutput {
	log.Println("Creating Stack with name: ", keyName(*stackInputParams.StackName))
	req := cf.Client.CreateStackRequest(stackInputParams)
	resp, err := req.Send()
	config.CheckError(err)
	return resp
}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code later on. By default FurnaceStack is used as name.`,
			Arguments:        "",
			Examples:         []string{""},
		},
	}
}
