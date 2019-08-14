package commands

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/codedeployiface"
	"github.com/go-furnace/go-furnace/handle"
)

type fakeDeleteAppCDClient struct {
	codedeployiface.ClientAPI
	err    error
	awsErr awserr.Error
}

func (fd *fakeDeleteAppCDClient) DeleteApplicationRequest(*codedeploy.DeleteApplicationInput) codedeploy.DeleteApplicationRequest {
	return codedeploy.DeleteApplicationRequest{
		Request: &aws.Request{
			Data:        &codedeploy.DeleteApplicationOutput{},
			Error:       fd.err,
			HTTPRequest: new(http.Request),
		},
	}
}

func TestGatherConfig(t *testing.T) {
	failed := false
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: nil, awsErr: nil}
	opts := &commander.CommandHelper{}
	gatherConfig(opts)
	if failed {
		t.Fatal("should not have failed")
	}
}

func TestGatherConfigCustomStack(t *testing.T) {
	failed := false
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: nil, awsErr: nil}
	opts := &commander.CommandHelper{}
	opts.Args = append(opts.Args, "fakeapp")
	gatherConfig(opts)
	if !failed {
		t.Fatal("should have failed")
	}
}

func TestDeletingApplication(t *testing.T) {
	failed := false
	var message string
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		if err, ok := a[0].(error); ok {
			message = err.Error()
		}
	}
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: nil, awsErr: nil}
	deleteApplication("fakeApp", client)
	if failed {
		t.Fatal("should not have called LogFatal. message was: ", message)
	}
}

func TestDeletingApplicationWithFailure(t *testing.T) {
	failed := false
	var message string
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		if err, ok := a[0].(error); ok {
			message = err.Error()
		}
	}
	client := new(CDClient)
	client.Client = &fakeDeleteAppCDClient{err: errors.New("test message"), awsErr: nil}
	deleteApplication("failedApp", client)
	if !failed {
		t.Fatal("should have called LogFatal")
	}
	if message != "test message" {
		t.Fatal("test message does not match expected: ", message)
	}
}

func TestDeleteAppCreate(t *testing.T) {
	wrapper := NewDeleteApp("furnace")
	if wrapper.Help.Arguments != "custom-config" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"", "custom-config"}) ||
		wrapper.Help.LongDescription != `Deletes a CodeDeploy Application complete with DeploymenyGroup and Deploys.` ||
		wrapper.Help.ShortDescription != "Deletes an Application" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
