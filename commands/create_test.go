package commands

import (
	"testing"

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

type fakeCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeCFClient) ValidateTemplate(input *cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	return &cloudformation.ValidateTemplateOutput{}, fc.err
}

func (fc *fakeCFClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return &cloudformation.CreateStackOutput{StackId: aws.String("DummyID")}, fc.err
	}
	return &cloudformation.CreateStackOutput{}, fc.err
}

func (fc *fakeCFClient) WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error {
	return fc.err
}

func (fc *fakeCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return notEmptyStack, fc.err
	}
	return &cloudformation.DescribeStacksOutput{}, fc.err
}

func TestCreateProcedure(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "NotEmptyStack"
	client.Client = &fakeCFClient{err: nil, stackname: stackname}
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
	client.Client = &fakeCFClient{err: nil, stackname: stackname}
	config := []byte("{}")
	stacks := create(stackname, config, client)
	if len(stacks) != 0 {
		t.Fatal("Stack was not empty: ", stacks)
	}
}
