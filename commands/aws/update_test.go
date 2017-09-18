package awscommands

import (
	"reflect"
	"testing"

	"errors"

	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	commander "github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeUpdateCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func init() {
	utils.LogFatalf = log.Fatalf
}

func (fc *fakeUpdateCFClient) ValidateTemplate(input *cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	if fc.stackname == "ValidationError" {
		return &cloudformation.ValidateTemplateOutput{}, fc.err
	}
	return &cloudformation.ValidateTemplateOutput{}, nil
}

func (fc *fakeUpdateCFClient) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return &cloudformation.UpdateStackOutput{StackId: aws.String("DummyID")}, fc.err
	}
	return &cloudformation.UpdateStackOutput{}, nil
}

func (fc *fakeUpdateCFClient) WaitUntilStackUpdateComplete(input *cloudformation.DescribeStacksInput) error {
	return nil
}

func (fc *fakeUpdateCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if fc.stackname == "NotEmptyStack" || fc.stackname == "DescribeStackFailed" {
		return NotEmptyStack, fc.err
	}
	return &cloudformation.DescribeStacksOutput{}, nil
}

func TestUpdateExecute(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeUpdateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	updateExecute(opts, client)
}

func TestUpdateExecuteEmptyStack(t *testing.T) {
	failed := false
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "EmptyStack"
	client.Client = &fakeUpdateCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	updateExecute(opts, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
}

func TestUpdateProcedure(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeUpdateCFClient{err: nil, stackname: stackname}
	config := []byte("{}")
	stacks := update(stackname, config, client)
	if len(stacks) == 0 {
		t.Fatal("Stack was not returned by create.")
	}
	if *stacks[0].StackName != "TestStack" {
		t.Fatal("Not the correct stack returned. Returned was:", stacks)
	}
}

func TestUpdateStackReturnsWithError(t *testing.T) {
	failed := false
	expectedMessage := "failed to create stack"
	var message string
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = a[0].(error).Error()
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeUpdateCFClient{err: errors.New(expectedMessage), stackname: stackname}
	config := []byte("{}")
	update(stackname, config, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
	if message != expectedMessage {
		t.Errorf("message did not equal expected message of '%s', was:%s", expectedMessage, message)
	}
}

func TestUpdateCreate(t *testing.T) {
	wrapper := NewUpdate("furnace")
	if wrapper.Help.Arguments != "" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{""}) ||
		wrapper.Help.LongDescription != `Update a stack with new parameters.` ||
		wrapper.Help.ShortDescription != "Update a stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
