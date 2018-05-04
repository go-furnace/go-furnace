package commands

import (
	"reflect"
	"testing"

	"github.com/Skarlso/go-furnace/config"
	awsconfig "github.com/Skarlso/go-furnace/furnace-aws/config"
	"github.com/Skarlso/go-furnace/handle"
	commander "github.com/Yitsushi/go-commander"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
)

type fakeDeleteCFClient struct {
	cloudformationiface.CloudFormationAPI
	stackname string
	err       error
}

func (fc *fakeDeleteCFClient) DeleteStackRequest(*cloudformation.DeleteStackInput) cloudformation.DeleteStackRequest {
	return cloudformation.DeleteStackRequest{
		Request: &aws.Request{
			Data: &cloudformation.DeleteStackOutput{},
		},
	}
}

func (fc *fakeDeleteCFClient) WaitUntilStackDeleteComplete(input *cloudformation.DescribeStacksInput) error {
	return fc.err
}

func TestDeleteProcedure(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	deleteStack(stackname, client)
}

func TestDeleteExecute(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	deleteExecute(opts, client)
}

func TestDeleteExecuteWithExtraStack(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	opts.Args = append(opts.Args, "teststack")
	deleteExecute(opts, client)
	if awsconfig.Config.Main.Stackname != "MyStack" {
		t.Fatal("test did not load the file requested.")
	}
}

func TestDeleteExecuteWithExtraStackNotFound(t *testing.T) {
	failed := false
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	opts.Args = append(opts.Args, "notfound")
	deleteExecute(opts, client)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
}

func TestPreDeletePlugins(t *testing.T) {
	ran := false
	runner := func(name string) {
		ran = true
	}
	plugins := awsconfig.RunPlugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[awsconfig.PREDELETE] = []awsconfig.RunPlugin{plugins}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	deleteExecute(opts, client)
	if !ran {
		t.Fatal("Predelete plugin was not executed.")
	}
}

func TestPostDeletePlugins(t *testing.T) {
	ran := false
	runner := func(name string) {
		ran = true
	}
	plugins := awsconfig.RunPlugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[awsconfig.POSTDELETE] = []awsconfig.RunPlugin{plugins}
	config.WAITFREQUENCY = 0
	client := new(CFClient)
	stackname := "ToDeleteStack"
	client.Client = &fakeDeleteCFClient{err: nil, stackname: stackname}
	opts := &commander.CommandHelper{}
	deleteExecute(opts, client)
	if !ran {
		t.Fatal("Postdelete plugin was not executed.")
	}
}

func TestDeleteCreate(t *testing.T) {
	wrapper := NewDelete("furnace")
	if wrapper.Help.Arguments != "custom-config" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"", "custom-config"}) ||
		wrapper.Help.LongDescription != `Delete a stack with a given name.` ||
		wrapper.Help.ShortDescription != "Delete a stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
