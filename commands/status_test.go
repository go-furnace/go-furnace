package commands

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeStatusCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeStatusCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if fc.stackname == "NotEmptyStack" {
		return NotEmptyStack, fc.err
	}
	return &cloudformation.DescribeStacksOutput{}, fc.err
}

func TestStatusCommandWithStackReturned(t *testing.T) {
	stackname := "NotEmptyStack"
	client := new(CFClient)
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	stacks := stackStatus(stackname, client)
	if len(stacks.Stacks) == 0 {
		t.Fatal("Zero stacks returned: ", stacks)
	}
}

func TestStatusWithNoStacks(t *testing.T) {
	stackname := "EmptyStacks"
	client := new(CFClient)
	client.Client = &fakeCreateCFClient{err: nil, stackname: stackname}
	stacks := stackStatus(stackname, client)
	if len(stacks.Stacks) != 0 {
		t.Fatal("Zero stacks returned: ", stacks)
	}
}
