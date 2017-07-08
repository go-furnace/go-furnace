package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	createExecute(opts, &client)
}

func createExecute(opts *commander.CommandHelper, client *CFClient) {
	stackname := config.STACKNAME
	template := config.LoadCFStackConfig()
	for _, p := range config.PluginRegistry[config.PRECREATE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
	stacks := create(stackname, template, client)
	for _, p := range config.PluginRegistry[config.POSTCREATE] {
		log.Println("Running plugin: ", p.Name)
		p.Run.(func())()
	}
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(*stacks[0].StackStatus))
	} else {
		utils.HandleFatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

var keyName = color.New(color.FgWhite, color.Bold).SprintFunc()

// create will create a full stack and encapsulate the functionality of
// the create command.
func create(stackname string, template []byte, cfClient *CFClient) []*cloudformation.Stack {
	validResp := cfClient.validateTemplate(template)
	stackParameters := gatherParameters(os.Stdin, validResp)
	stackInputParams := &cloudformation.CreateStackInput{
		StackName: aws.String(stackname),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
		},
		Parameters:   stackParameters,
		TemplateBody: aws.String(string(template)),
	}
	resp := cfClient.createStack(stackInputParams)
	log.Println("Create stack response: ", resp.GoString())
	cfClient.waitForStackStatus(stackname, "CREATE_COMPLETE")
	descResp := cfClient.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	return descResp.Stacks
}

func (cf *CFClient) createStack(stackInputParams *cloudformation.CreateStackInput) *cloudformation.CreateStackOutput {
	log.Println("Creating Stack with name: ", keyName(*stackInputParams.StackName))
	resp, err := cf.Client.CreateStack(stackInputParams)
	utils.CheckError(err)
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
