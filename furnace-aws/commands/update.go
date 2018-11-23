package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/handle"
	"github.com/satori/go.uuid"
)

// Update command.
type Update struct {
}

// Execute defines what this command does.
func (c *Update) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	update(opts, &client)
}

func update(opts *commander.CommandHelper, client *CFClient) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	stackname := awsconfig.Config.Main.Stackname
	template := awsconfig.LoadCFStackConfig()
	changeSetName := createChangeSet(stackname, template, client)
	client.waitForChangeSetToBeApplied(stackname, changeSetName)
	describeChangeInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: &changeSetName,
		StackName:     &stackname,
	}
	changes := client.Client.DescribeChangeSetRequest(describeChangeInput)
	resp, _ := changes.Send()
	// Get confirm for applying update.

	//reading a string
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Would you like to apply the changes? (y/N):")
	confirm, _ := reader.ReadString('\n')

	if confirm != "y" {
		log.Println("Cancelling without applying change set.")
		return
	}

	executeChangeInput := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      resp.ChangeSetName,
		ClientRequestToken: resp.NextToken,
		StackName:          &stackname,
	}
	executeChangeRequest := client.Client.ExecuteChangeSetRequest(&executeChangeInput)
	executeChangeRequest.Send()
	client.waitForStackUpdateComplete(stackname)
	descResp := client.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	stacks := descResp.Stacks
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(stacks[0].StackStatus))
	} else {
		handle.Fatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

func createChangeSet(stackname string, template []byte, cfClient *CFClient) string {
	changeSetName := uuid.NewV4()
	validResp := cfClient.validateTemplate(template)
	stackParameters := gatherParameters(os.Stdin, validResp)
	changeSetRequestInput := &cloudformation.CreateChangeSetInput{
		StackName: aws.String(stackname),
		Capabilities: []cloudformation.Capability{
			cloudformation.CapabilityCapabilityIam,
		},
		Parameters:    stackParameters,
		TemplateBody:  aws.String(string(template)),
		ChangeSetName: aws.String(changeSetName.String()),
		ChangeSetType: cloudformation.ChangeSetTypeUpdate,
	}
	changeSetRequest := cfClient.Client.CreateChangeSetRequest(changeSetRequestInput)
	changeSetRequest.Send()
	return changeSetName.String()
}

func (cf *CFClient) waitForChangeSetToBeApplied(stackname, changeSetName string) {
	describeChangeInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: &changeSetName,
		StackName:     &stackname,
	}
	waitForFunctionWithStatusOutput("UPDATE_COMPLETE", config.WAITFREQUENCY, func() {
		cf.Client.WaitUntilChangeSetCreateComplete(describeChangeInput)
	})
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
	req := cf.Client.UpdateStackRequest(stackInputParams)
	resp, err := req.Send()
	handle.Error(err)
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
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
