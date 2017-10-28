package awscommands

import (
	"fmt"
	"log"

	awsconfig "github.com/Skarlso/go-furnace/config/aws"
	config "github.com/Skarlso/go-furnace/config/common"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
)

// Status command.
type Status struct {
}

// ResourceStatus defines a resource.
type ResourceStatus struct {
	// Status is the status of the resource
	Status string
	// PhysicalId of the resource
	PhysicalID string
	// LogicalId of the resource
	LogicalID string
	// Type of the resource
	Type string
}

// Execute defines what this command does.
func (c *Status) Execute(opts *commander.CommandHelper) {
	stackname := config.STACKNAME
	sess := session.New(&aws.Config{Region: aws.String(awsconfig.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	stack := stackStatus(stackname, &client)
	info := color.New(color.FgWhite, color.Bold).SprintFunc()
	log.Println("Stack state is: ", info(stack.Stacks[0].GoString()))
	stackResources := stackResources(stackname, &client)
	printStackResources(stackResources)
}

func stackStatus(stackname string, cfClient *CFClient) *cloudformation.DescribeStacksOutput {
	descResp, err := cfClient.Client.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	utils.CheckError(err)
	fmt.Println()
	return descResp
}

func stackResources(stackname string, cfClient *CFClient) []ResourceStatus {
	resources := make([]ResourceStatus, 0)
	descResp, err := cfClient.Client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{StackName: aws.String(stackname)})
	utils.CheckError(err)
	for _, r := range descResp.StackResources {
		res := ResourceStatus{Status: *r.ResourceStatus, PhysicalID: *r.PhysicalResourceId, LogicalID: *r.LogicalResourceId, Type: *r.ResourceType}
		resources = append(resources, res)
	}
	fmt.Println()
	return resources
}

func printStackResources(resources []ResourceStatus) {
	info := color.New(color.FgWhite, color.Bold).SprintFunc()
	fmt.Println(info("___________________"))
	for _, r := range resources {
		fmt.Print(info(r))
	}
	fmt.Println()
}

func (r ResourceStatus) String() string {
	var red = color.New(color.FgRed).SprintFunc()
	var yellow = color.New(color.FgYellow).SprintFunc()
	ret := ""
	ret += fmt.Sprintf("|Name:          %s|\n|Id:            %s|\n|Status:        %s|\n|Type:          %s|\n",
		red(r.LogicalID),
		yellow(r.PhysicalID),
		yellow(r.Status),
		yellow(r.Type))
	ret += "-------------------\n"
	return ret
}

// NewStatus Creates a new Status command.
func NewStatus(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Status{},
		Help: &commander.CommandDescriptor{
			Name:             "status",
			ShortDescription: "Status of a stack.",
			LongDescription:  `Get detailed status of the stack.`,
			Arguments:        "",
			Examples:         []string{""},
		},
	}
}
