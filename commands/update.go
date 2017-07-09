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

// Update command.
type Update struct {
}

// Execute defines what this command does.
func (c *Update) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	updateExecute(opts, &client)
}

func updateExecute(opts *commander.CommandHelper, client *CFClient) {
	stackname := config.STACKNAME
	template := config.LoadCFStackConfig()
	stacks := update(stackname, template, client)
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(*stacks[0].StackStatus))
	} else {
		utils.HandleFatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
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
	utils.WaitForFunctionWithStatusOutput("UPDATE_COMPLETE", config.WAITFREQUENCY, func() {
		cf.Client.WaitUntilStackUpdateComplete(describeStackInput)
	})
}

func (cf *CFClient) updateStack(stackInputParams *cloudformation.UpdateStackInput) *cloudformation.UpdateStackOutput {
	log.Println("Updating Stack with name: ", keyName(*stackInputParams.StackName))
	resp, err := cf.Client.UpdateStack(stackInputParams)
	utils.CheckError(err)
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