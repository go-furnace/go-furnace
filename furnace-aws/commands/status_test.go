package commands

import (
	"testing"
)

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
