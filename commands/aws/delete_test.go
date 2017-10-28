package awscommands

import (
	"reflect"
	"testing"

	awsconfig "github.com/Skarlso/go-furnace/config/aws"
	config "github.com/Skarlso/go-furnace/config/common"
	commander "github.com/Yitsushi/go-commander"
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

func TestPreDeletePlugins(t *testing.T) {
	ran := false
	runner := func() {
		ran = true
	}
	plugins := awsconfig.Plugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[config.PREDELETE] = []awsconfig.Plugin{plugins}
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
	runner := func() {
		ran = true
	}
	plugins := awsconfig.Plugin{
		Name: "testPlugin",
		Run:  runner,
	}
	awsconfig.PluginRegistry[config.POSTDELETE] = []awsconfig.Plugin{plugins}
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
	if wrapper.Help.Arguments != "" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{""}) ||
		wrapper.Help.LongDescription != `Delete a stack with a given name.` ||
		wrapper.Help.ShortDescription != "Delete a stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
