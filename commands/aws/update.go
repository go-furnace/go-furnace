package awscommands

import (
	"fmt"
	"log"
	"os"

	awsconfig "github.com/Skarlso/go-furnace/config/aws"
	config "github.com/Skarlso/go-furnace/config/common"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
)

// Update command.
type Update struct {
}

// Execute defines what this command does.
func (c *Update) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String(awsconfig.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	updateExecute(opts, &client)
}

func updateExecute(opts *commander.CommandHelper, client *CFClient) {
	stackname := config.STACKNAME
	template := awsconfig.LoadCFStackConfig()
	stacks := update(stackname, template, client)
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(*stacks[0].StackStatus))
	} else {
		config.HandleFatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

func update(stackname string, template []byte, cfClient *CFClient) []*cloudformation.Stack {
	validResp := cfClient.validateTemplate(template)
	stackParameters := gatherParameters(os.Stdin, validResp)
	stackInputParams := &cloudformation.UpdateStackInput{
		StackName: aws.String(stackname),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
		},
		Parameters:   stackParameters,
		TemplateBody: aws.String(string(template)),
	}
	resp := cfClient.updateStack(stackInputParams)
	log.Println("Update stack response: ", resp.GoString())
	cfClient.waitForStackUpdateComplete(stackname)
	descResp := cfClient.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	return descResp.Stacks
}

func (cf *CFClient) waitForStackUpdateComplete(stackname string) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("UPDATE_COMPLETE", config.WAITFREQUENCY, func() {
		cf.Client.WaitUntilStackUpdateComplete(describeStackInput)
	})
}

func (cf *CFClient) updateStack(stackInputParams *cloudformation.UpdateStackInput) *cloudformation.UpdateStackOutput {
	log.Println("Updating Stack with name: ", keyName(*stackInputParams.StackName))
	resp, err := cf.Client.UpdateStack(stackInputParams)
	config.CheckError(err)
	return resp
}

// NewUpdate Updates a new Update command.
func NewUpdate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Update{},
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update a stack",
			LongDescription:  `Update a stack with new parameters.`,
			Arguments:        "",
			Examples:         []string{""},
		},
	}
}
