package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/handle"
	"github.com/google/uuid"
)

// Update command.
type Update struct {
	client *CFClient
}

// DescribeChangeSetRequestSender describes a sender interface in order to mock
// a support call to AWS directly from the constructed request object.
type DescribeChangeSetRequestSender interface {
	Send(context.Context) (*cloudformation.DescribeChangeSetResponse, error)
}

// ExecuteChangeSetRequestSender describes a sender interface in order to mock
// a support call to AWS directly from the constructed request object.
type ExecuteChangeSetRequestSender interface {
	Send(context.Context) (*cloudformation.ExecuteChangeSetResponse, error)
}

// Execute defines what this command does.
func (u *Update) Execute(opts *commander.CommandHelper) {
	override := opts.Flag("y")
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	stackname := awsconfig.Config.Main.Stackname
	template := awsconfig.LoadCFStackConfig()
	changeSetName := createChangeSet(stackname, template, u.client)
	if changeSetName == "" {
		handle.Fatal("Change set name was empty.", errors.New("change set was empty"))
	}
	u.client.waitForChangeSetToBeApplied(stackname, changeSetName)
	describeChangeInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: &changeSetName,
		StackName:     &stackname,
	}
	changes := u.client.Client.DescribeChangeSetRequest(describeChangeInput)
	resp, err := sendDescribeChangeSetRequest(changes)
	handle.Error(err)

	if resp == nil {
		log.Println("describe change set request send returned nil")
		return
	}

	for i, change := range resp.Changes {
		fmt.Printf("=====  Begin Change Number %s =====\n", keyName(i))
		fmt.Println(change.ResourceChange.String())
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
	executeChangeRequest := u.client.Client.ExecuteChangeSetRequest(&executeChangeInput)
	_, err = sendExecuteChangeSetRequestSender(executeChangeRequest)
	handle.Error(err)
	u.client.waitForStackUpdateComplete(stackname)
	descResp := u.client.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	stacks := descResp.Stacks
	var red = color.New(color.FgRed).SprintFunc()
	if stacks != nil {
		log.Println("Stack state is: ", red(stacks[0].StackStatus))
	} else {
		handle.Fatal(fmt.Sprintf("No stacks found with name: %s", keyName(stackname)), nil)
	}
}

func sendDescribeChangeSetRequest(send DescribeChangeSetRequestSender) (*cloudformation.DescribeChangeSetResponse, error) {
	return send.Send(context.Background())
}

func sendExecuteChangeSetRequestSender(send ExecuteChangeSetRequestSender) (*cloudformation.ExecuteChangeSetResponse, error) {
	return send.Send(context.Background())
}

func createChangeSet(stackname string, template []byte, cfClient *CFClient) string {
	changeSetName, _ := uuid.NewUUID()
	validResp := cfClient.validateTemplate(template)
	if validResp == nil {
		log.Println("The response from AWS to validate was nil.")
		return ""
	}
	stackParameters := gatherParameters(os.Stdin, validResp.ValidateTemplateOutput)
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
	_, err := changeSetRequest.Send(context.Background())
	handle.Error(err)
	return changeSetName.String()
}

func (cf *CFClient) waitForChangeSetToBeApplied(stackname, changeSetName string) {
	describeChangeInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: &changeSetName,
		StackName:     &stackname,
	}
	waitForFunctionWithStatusOutput("UPDATE_COMPLETE", config.WAITFREQUENCY, func() {
		_ = cf.Client.WaitUntilChangeSetCreateComplete(context.Background(), describeChangeInput)
	})
}

func (cf *CFClient) waitForStackUpdateComplete(stackname string) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	waitForFunctionWithStatusOutput("UPDATE_COMPLETE", config.WAITFREQUENCY, func() {
		err := cf.Client.WaitUntilStackUpdateComplete(context.Background(), describeStackInput)
		if err != nil {
			return
		}
	})
}

// NewUpdate Updates a new Update command.
func NewUpdate(appName string) *commander.CommandWrapper {
	log.Println("Creating cloud formation session.")
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	u := Update{
		client: &client,
	}
	return &commander.CommandWrapper{
		Handler: &u,
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update a stack",
			LongDescription:  `Update a stack with new parameters. -y can be given to automatically accept the applying of a changeset.`,
			Arguments:        "custom-config [-y]",
			Examples:         []string{"", "custom-config", "-y", "mystack -y"},
		},
	}
}
