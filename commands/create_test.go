package commands

import (
	"testing"

	"github.com/Skarlso/go-furnace/config"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeCFClient struct {
	cloudformationiface.CloudFormationAPI
	err error
}

func (fc *fakeCFClient) ValidateTemplate(input *cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	return &cloudformation.ValidateTemplateOutput{}, fc.err
}

func (fc *fakeCFClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return &cloudformation.CreateStackOutput{}, fc.err
}

func (fc *fakeCFClient) WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error {
	return fc.err
}

func (fc *fakeCFClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{}, fc.err
}

func TestCreateCommand(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	client.Client = &fakeCFClient{err: nil}
	config := []byte("{}")
	stackname := "TestStack"
	create(stackname, config, client)
}
