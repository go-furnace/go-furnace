package awscommands

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"reflect"

	"errors"

	"log"

	awsconfig "github.com/Skarlso/go-furnace/config/aws"
	config "github.com/Skarlso/go-furnace/config/common"
	commander "github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeCreateCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func init() {
	config.LogFatalf = log.Fatalf
}

func (fc *fakeCreateCFClient) ValidateTemplate(input *cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	if fc.stackname == "ValidationError" {
		return &cloudformation.ValidateTemplateOutput{}, fc.err
	}
	return &cloudformation.ValidateTemplateOutput{}, nil
}

func (fc *fakeCreateCFClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return &cloudformation.CreateStackOutput{StackId: aws.String("DummyID")}, fc.err
	}
	return &cloudformation.CreateStackOutput{}, nil
}

func (fc *fakeCreateCFClient) WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error {
	return nil
}

func (fc *fakeCreateCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if fc.stackname == "NotEmptyStack" || fc.stackname == "DescribeStackFailed" {
		return NotEmptyStack, fc.err
	}
	return &cloudformation.DescribeStacksOutput{}, nil
}

func TestCreateExecute(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	createExecute(opts, client)
}

func TestCreateExecuteEmptyStack(t *testing.T) {
	failed := false
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "EmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	createExecute(opts, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
}

func TestCreateProcedure(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	config := []byte("{}")
	stacks := create(stackname, config, client)
	if len(stacks) == 0 {
		t.Fatal("Stack was not returned by create.")
	}
	if *stacks[0].StackName != "TestStack" {
		t.Fatal("Not the correct stack returned. Returned was:", stacks)
	}
}

func TestCreateStackReturnsWithError(t *testing.T) {
	failed := false
	expectedMessage := "failed to create stack"
	var message string
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = a[0].(error).Error()
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCreateCFClient{err: errors.New(expectedMessage), stackname: stackname}
	config := []byte("{}")
	create(stackname, config, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
	if message != expectedMessage {
		t.Errorf("message did not equal expected message of '%s', was:%s", expectedMessage, message)
	}
}

func TestDescribeStackReturnsWithError(t *testing.T) {
	failed := false
	var message string
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		if err, ok := a[0].(error); ok {
			message = err.Error()
		}
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "DescribeStackFailed"
	client.Client = &fakeCreateCFClient{err: errors.New("failed describe stack"), stackname: stackname}
	config := []byte("{}")
	create(stackname, config, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
	if message != "failed describe stack" {
		t.Error("message did not equal expected message of 'failed describe stack', was:", message)
	}
}

func TestValidateReturnsWithError(t *testing.T) {
	failed := false
	expectedMessage := "validation error occurred"
	var message string
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		if err, ok := a[0].(error); ok {
			message = err.Error()
		}
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ValidationError"
	client.Client = &fakeCreateCFClient{err: errors.New(expectedMessage), stackname: stackname}
	config := []byte("{}")
	create(stackname, config, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
	if message != expectedMessage {
		t.Errorf("message did not equal expected message of '%s', was:%s", expectedMessage, message)
	}
}

func TestCreateReturnsEmptyStack(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "EmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	config := []byte("{}")
	stacks := create(stackname, config, client)
	if len(stacks) != 0 {
		t.Fatal("Stack was not empty: ", stacks)
	}
}

func TestGatheringParametersWithoutSpecifyingUserInputShouldUseDefaultValue(t *testing.T) {
	in, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	validOutput := &cloudformation.ValidateTemplateOutput{
		Parameters: []*cloudformation.TemplateParameter{
			{
				DefaultValue: aws.String("DefaultValue"),
				Description:  aws.String("Description"),
				NoEcho:       aws.Bool(false),
				ParameterKey: aws.String("Key"),
			},
		},
	}
	params := gatherParameters(in, validOutput)
	if *params[0].ParameterKey != "Key" {
		t.Fatal("Key did not equal expected key. Was:", *params[0].ParameterKey)
	}
	if *params[0].ParameterValue != "DefaultValue" {
		t.Fatal("Value did not equal expected value. Was:", *params[0].ParameterValue)
	}
}

func TestGatheringParametersWithUserInputShouldUseInput(t *testing.T) {
	// Create a temp file
	in, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	// Write the new value in that file
	_, err = io.WriteString(in, "NewValue\n")
	if err != nil {
		t.Fatal(err)
	}
	// Set the starting point for the next read to be the beginning of the file
	_, err = in.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}
	// Setup the input
	validOutput := &cloudformation.ValidateTemplateOutput{
		Parameters: []*cloudformation.TemplateParameter{
			{
				DefaultValue: aws.String("DefaultValue"),
				Description:  aws.String("Description"),
				NoEcho:       aws.Bool(false),
				ParameterKey: aws.String("Key"),
			},
		},
	}
	params := gatherParameters(in, validOutput)
	if *params[0].ParameterKey != "Key" {
		t.Fatal("Key did not equal expected key. Was:", *params[0].ParameterKey)
	}
	if *params[0].ParameterValue != "NewValue" {
		t.Fatal("Value did not equal expected value. Was:", *params[0].ParameterValue)
	}
}

func TestNewCreate(t *testing.T) {
	wrapper := NewCreate("furnace")
	if wrapper.Help.Arguments != "" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{""}) ||
		wrapper.Help.LongDescription != `Create a stack on which to deploy code later on. By default FurnaceStack is used as name.` ||
		wrapper.Help.ShortDescription != "Create a stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}

func TestPreCreatePlugins(t *testing.T) {
	ran := false
	runner := func() {
		ran = true
	}
	plugins := awsconfig.Plugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[config.PRECREATE] = []awsconfig.Plugin{plugins}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	createExecute(opts, client)
	if !ran {
		t.Fatal("Precreate plugin was not executed.")
	}
}

func TestPostCreatePlugins(t *testing.T) {
	ran := false
	runner := func() {
		ran = true
	}
	plugins := awsconfig.Plugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[config.POSTCREATE] = []awsconfig.Plugin{plugins}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	createExecute(opts, client)
	if !ran {
		t.Fatal("Postcreate plugin was not executed.")
	}
}
