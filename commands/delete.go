package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Delete command.
type Delete struct {
}

// Execute defines what this command does.
func (c *Delete) Execute(opts *commander.CommandHelper) {

	stackname := opts.Arg(0)
	if len(stackname) < 1 {
		log.Fatalln("A stackname to delete must be provided.")
	}

	log.Println("Deleting CloudFormation stack with name:", stackname)
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cfClient := cloudformation.New(sess, nil)
	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackname),
	}
	resp, err := cfClient.DeleteStack(params)
	utils.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	log.Println("Delete stack response: ", resp.GoString())
	utils.WaitForFunctionWithStatusOutput("DELETE_COMPLETE", func() {
		cfClient.WaitUntilStackDeleteComplete(describeStackInput)
	})
}

// NewDelete Creates a new Delete command.
func NewDelete(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Delete{},
		Help: &commander.CommandDescriptor{
			Name:             "delete",
			ShortDescription: "Delete a stack",
			LongDescription:  `Delete a stack with a given name.`,
			Arguments:        "name",
			Examples:         []string{"delete FurnaceStack"},
		},
	}
}
