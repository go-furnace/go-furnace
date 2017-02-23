package commands

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"reflect"

	"github.com/Skarlso/go-furnace/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

var notEmptyStack = &cloudformation.DescribeStacksOutput{
	Stacks: []*cloudformation.Stack{
		&cloudformation.Stack{StackName: aws.String("TestStack")},
	},
}

type fakeCreateCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeCreateCFClient) ValidateTemplate(input *cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	return &cloudformation.ValidateTemplateOutput{}, fc.err
}

func (fc *fakeCreateCFClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return &cloudformation.CreateStackOutput{StackId: aws.String("DummyID")}, fc.err
	}
	return &cloudformation.CreateStackOutput{}, fc.err
}

func (fc *fakeCreateCFClient) WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error {
	return fc.err
}

func (fc *fakeCreateCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return notEmptyStack, fc.err
	}
	return &cloudformation.DescribeStacksOutput{}, fc.err
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
			&cloudformation.TemplateParameter{
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
			&cloudformation.TemplateParameter{
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
	if wrapper.Help.Arguments != "name" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"create", "create MyStackName"}) ||
		wrapper.Help.LongDescription != `Create a stack on which to deploy code to later on. By default FurnaceStack is used as name.` ||
		wrapper.Help.ShortDescription != "Create a stack" {
		t.Fatal("wrapper did not match with given params")
	}
}
