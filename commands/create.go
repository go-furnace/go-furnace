package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Create command.
type Create struct {
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	stackname := opts.Arg(0)
	if len(stackname) < 1 {
		stackname = config.STACKNAME
	}

	config := config.LoadCFStackConfig()
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cfClient := cloudformation.New(sess, nil)
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(config)),
	}

	m := make(map[string]interface{})
	json.Unmarshal(config, &m)
	validResp, err := cfClient.ValidateTemplate(validateParams)
	log.Println("Response from validate:", validResp)
	utils.CheckError(err)
	for k, v := range validResp.Parameters {
		log.Println("k;v", k, *v.DefaultValue)
	}
	// os.Exit(1)

	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackname),
		TemplateBody: aws.String(string(config)),
	}
	resp, err := cfClient.CreateStack(stackInputParams)
	utils.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	log.Println("Create stack response: ", resp.GoString())
	utils.WaitForFunctionWithStatusOutput("CREATE_COMPLETE", func() {
		cfClient.WaitUntilStackCreateComplete(describeStackInput)
	})
	descResp, err := cfClient.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	utils.CheckError(err)
	fmt.Println()
	log.Println("Stack state is: ", *descResp.Stacks[0].StackStatus)

}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code to later on. By default FurnaceStack is used as name.`,
			Arguments:        "name",
			Examples:         []string{"create", "create MyStackName"},
		},
	}
}
