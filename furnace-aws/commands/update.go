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
	uuid "github.com/google/uuid"
)

// Update command.
type Update struct {
}

// DescribeChangeSetRequestSender describes a sender interface in order to mock
// a support call to AWS directly from the constructed request object.
type DescribeChangeSetRequestSender interface {
	Send() (*cloudformation.DescribeChangeSetOutput, error)
}

// ExecuteChangeSetRequestSender describes a sender interface in order to mock
// a support call to AWS directly from the constructed request object.
type ExecuteChangeSetRequestSender interface {
	Send() (*cloudformation.ExecuteChangeSetOutput, error)
}

// Execute defines what this command does.
func (c *Update) Execute(opts *commander.CommandHelper) {
	log.Println("Creating cloud formation session.")
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	override := opts.Flag("y")
	update(opts, &client, override)
}

// Todo the CFClient needs an inner property
// DescribeSender
// ExecuteSender
// And use those in the sender parameter.
// Block that later on.
func update(opts *commander.CommandHelper, client *CFClient, override bool) {
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
	resp, err := sendDescribeChangeSetRequest(changes)
	handle.Error(err)

	if resp == nil {
		log.Println("describe change set request send returned nil")
		return
	}

	for i, change := range resp.Changes {
		fmt.Printf("=====  Begin Change Number %s =====\n", keyName(i))
		fmt.Println(change.ResourceChange.GoString())
		fmt.Printf("===== End of Change Number %s =====\n", keyName(i))
	}

	// Get confirm for applying update.
	if !override {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Would you like to apply the changes? (y/N):")
		confirm, _ := reader.ReadString('\n')
		if confirm != "y" {
			log.Println("Cancelling without applying change set.")
			return
		}
	}

	executeChangeInput := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      resp.ChangeSetName,
		ClientRequestToken: resp.NextToken,
		StackName:          &stackname,
	}
	executeChangeRequest := client.Client.ExecuteChangeSetRequest(&executeChangeInput)
	sendExecuteChangeSetRequestSender(executeChangeRequest)
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

func sendDescribeChangeSetRequest(send DescribeChangeSetRequestSender) (*cloudformation.DescribeChangeSetOutput, error) {
	return send.Send()
}

func sendExecuteChangeSetRequestSender(send ExecuteChangeSetRequestSender) (*cloudformation.ExecuteChangeSetOutput, error) {
	return send.Send()
}

func createChangeSet(stackname string, template []byte, cfClient *CFClient) string {
	changeSetName, _ := uuid.NewUUID()
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
			LongDescription:  `Update a stack with new parameters. -y can be given to automatically accept the applying of a changeset.`,
			Arguments:        "custom-config [-y]",
			Examples:         []string{"", "custom-config", "-y", "mystack -y"},
		},
	}
}
