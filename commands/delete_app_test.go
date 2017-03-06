package commands

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codedeploy/codedeployiface"
)

type fakeDeleteAppCDClient struct {
	codedeployiface.CodeDeployAPI
	err    error
	awsErr awserr.Error
}

func (fd *fakeDeleteAppCDClient) DeleteApplication(*codedeploy.DeleteApplicationInput) (*codedeploy.DeleteApplicationOutput, error) {
	return &codedeploy.DeleteApplicationOutput{}, fd.err
}

func TestDeletingApplication(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: nil, awsErr: nil}
	deleteApplication("fakeApp", client)
}
