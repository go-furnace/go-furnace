package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
	validResp, err := cfClient.ValidateTemplate(validateParams)
	// log.Println("Response from validate:", validResp)
	utils.CheckError(err)
	var stackParameters []*cloudformation.Parameter
	keyName := color.New(color.FgWhite, color.Bold).SprintFunc()
	defaultValue := color.New(color.FgHiBlack, color.Italic).SprintFunc()
	for _, v := range validResp.Parameters {
		var param cloudformation.Parameter
		fmt.Printf("%s - '%s'(%s):", *v.Description, keyName(*v.ParameterKey), defaultValue(*v.DefaultValue))
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		param.SetParameterKey(*v.ParameterKey)
		text = strings.Trim(text, "\n")
		if len(text) > 0 {
			param.SetParameterValue(*aws.String(text))
		} else {
			param.SetParameterValue(*v.DefaultValue)
		}
		stackParameters = append(stackParameters, &param)
	}

	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackname),
		Parameters:   stackParameters,
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
	var red = color.New(color.FgRed).SprintFunc()
	log.Println("Stack state is: ", red(*descResp.Stacks[0].StackStatus))
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
