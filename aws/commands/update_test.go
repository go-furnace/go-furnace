package commands

import (
	"reflect"
	"testing"

	"errors"

	"log"

	config "github.com/Skarlso/go-furnace/config"
	commander "github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
)

type fakeUpdateCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func init() {
	config.LogFatalf = log.Fatalf
}

func (fc *fakeUpdateCFClient) ValidateTemplateRequest(input *cloudformation.ValidateTemplateInput) cloudformation.ValidateTemplateRequest {
	return cloudformation.ValidateTemplateRequest{
		Request: &aws.Request{
			Data:  &cloudformation.ValidateTemplateOutput{},
			Error: fc.err,
		},
		Input: input,
	}
}

func (fc *fakeUpdateCFClient) UpdateStackRequest(input *cloudformation.UpdateStackInput) cloudformation.UpdateStackRequest {
	return cloudformation.UpdateStackRequest{
		Request: &aws.Request{
			Data: &cloudformation.UpdateStackOutput{
				StackId: aws.String("DummyID"),
			},
			Error: fc.err,
		},
		Input: input,
	}
}

func (fc *fakeUpdateCFClient) WaitUntilStackUpdateComplete(input *cloudformation.DescribeStacksInput) error {
	return nil
}

func (fc *fakeUpdateCFClient) DescribeStacksRequest(input *cloudformation.DescribeStacksInput) cloudformation.DescribeStacksRequest {
	if fc.stackname == "NotEmptyStack" {
		return cloudformation.DescribeStacksRequest{
			Request: &aws.Request{
				Data:  &NotEmptyStack,
				Error: fc.err,
			},
		}
	}
	return cloudformation.DescribeStacksRequest{
		Request: &aws.Request{
			Data: &cloudformation.DescribeStacksOutput{},
		},
	}
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
	config.LogFatalf = func(s string, a ...interface{}) {
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
	config.LogFatalf = func(s string, a ...interface{}) {
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
