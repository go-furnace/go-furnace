package commands

import (
	"testing"

	"github.com/Skarlso/go-furnace/config"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeDeleteCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeDeleteCFClient) DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	return &cloudformation.DeleteStackOutput{}, fc.err
}

func (fc *fakeCreateCFClient) WaitUntilStackDeleteComplete(input *cloudformation.DescribeStacksInput) error {
	return fc.err
}

func TestDeleteProcedure(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	// deleteStack(stackname, client)
}
