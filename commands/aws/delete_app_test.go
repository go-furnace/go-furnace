package awscommands

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/codedeployiface"
)

type fakeDeleteAppCDClient struct {
	codedeployiface.CodeDeployAPI
	err    error
	awsErr awserr.Error
}

func (fd *fakeDeleteAppCDClient) DeleteApplicationRequest(*codedeploy.DeleteApplicationInput) codedeploy.DeleteApplicationRequest {
	return codedeploy.DeleteApplicationRequest{
		Request: &aws.Request{
			Data: &codedeploy.DeleteApplicationOutput{},
		},
	}
}

func TestDeletingApplication(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: nil, awsErr: nil}
	deleteApplication("fakeApp", client)
}

func TestDeleteAppCreate(t *testing.T) {
	wrapper := NewDeleteApp("furnace")
	if wrapper.Help.Arguments != "name" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"delete-application", "delete-application CustomApplicationName"}) ||
		wrapper.Help.LongDescription != `Deletes a CodeDeploy Application complete with DeploymenyGroup and Deploys.` ||
		wrapper.Help.ShortDescription != "Deletes an Application" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
