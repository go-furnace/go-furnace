package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/plugins"
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
	stackname := config.STACKNAME
	template := config.LoadCFStackConfig()
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	preCreatePlugins := plugins.GetPluginsForEvent(config.PRECREATE)
	log.Println("The following plugins will be triggered pre-create: ", preCreatePlugins)
	for _, p := range preCreatePlugins {
		p.RunPlugin()
	}
	stacks := create(stackname, template, &client)
	postCreatePlugins := plugins.GetPluginsForEvent(config.POSTCREATE)
	log.Println("The following plugins will be triggered post-create: ", postCreatePlugins)
	for _, p := range postCreatePlugins {
		p.RunPlugin()
	}
	var red = color.New(color.FgRed).SprintFunc()
	if len(stacks) > 0 {
		log.Println("Stack state is: ", red(*stacks[0].StackStatus))
	} else {
		log.Fatalln("No stacks found with name: ", keyName(stackname))
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
	cfClient.waitForStackComplete(stackname)
	descResp := cfClient.describeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	fmt.Println()
	return descResp.Stacks
}

func gatherParameters(source *os.File, params *cloudformation.ValidateTemplateOutput) []*cloudformation.Parameter {
	var stackParameters []*cloudformation.Parameter
	defaultValue := color.New(color.FgHiBlack, color.Italic).SprintFunc()
	log.Println("Gathering parameters.")
	for _, v := range params.Parameters {
		var param cloudformation.Parameter
		fmt.Printf("%s - '%s'(%s):", *v.Description, keyName(*v.ParameterKey), defaultValue(*v.DefaultValue))
		text := readInputFrom(source)
		param.SetParameterKey(*v.ParameterKey)
		text = strings.Trim(text, "\n")
		if len(text) > 0 {
			param.SetParameterValue(*aws.String(text))
		} else {
			param.SetParameterValue(*v.DefaultValue)
		}
		stackParameters = append(stackParameters, &param)
	}
	return stackParameters
}

func readInputFrom(source *os.File) string {
	reader := bufio.NewReader(source)
	text, _ := reader.ReadString('\n')
	return text
}

func (cf *CFClient) waitForStackComplete(stackname string) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	utils.WaitForFunctionWithStatusOutput("CREATE_COMPLETE", config.WAITFREQUENCY, func() {
		cf.Client.WaitUntilStackCreateComplete(describeStackInput)
	})
}

func (cf *CFClient) createStack(stackInputParams *cloudformation.CreateStackInput) *cloudformation.CreateStackOutput {
	log.Println("Creating Stack with name: ", keyName(*stackInputParams.StackName))
	resp, err := cf.Client.CreateStack(stackInputParams)
	utils.CheckError(err)
	return resp
}

func (cf *CFClient) describeStacks(descStackInput *cloudformation.DescribeStacksInput) *cloudformation.DescribeStacksOutput {
	descResp, err := cf.Client.DescribeStacks(descStackInput)
	utils.CheckError(err)
	return descResp
}

func (cf *CFClient) validateTemplate(template []byte) *cloudformation.ValidateTemplateOutput {
	log.Println("Validating template.")
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(template)),
	}
	resp, err := cf.Client.ValidateTemplate(validateParams)
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
			LongDescription:  `Create a stack on which to deploy code to later on. By default FurnaceStack is used as name.`,
			Arguments:        "",
			Examples:         []string{"create"},
		},
	}
}
