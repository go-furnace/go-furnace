package commands

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
)

type fakeStatusCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeStatusCFClient) DescribeStacksRequest(input *cloudformation.DescribeStacksInput) cloudformation.DescribeStacksRequest {
	if fc.stackname == "NotEmptyStack" {
		return cloudformation.DescribeStacksRequest{
			Request: &aws.Request{
				Data:  NotEmptyStack,
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
