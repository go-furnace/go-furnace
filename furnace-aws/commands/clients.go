package commands

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/codedeployiface"
	"github.com/aws/aws-sdk-go-v2/service/iam/iamiface"
	"github.com/go-furnace/go-furnace/handle"
	"log"
)

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.ClientAPI
}

func (cf *CFClient) describeStacks(descStackInput *cloudformation.DescribeStacksInput) *cloudformation.DescribeStacksOutput {
	req := cf.Client.DescribeStacksRequest(descStackInput)
	descResp, err := req.Send(context.Background())
	handle.Error(err)
	if descResp == nil {
		handle.Fatal("the response was nil: ", errors.New("the response was nil"))
		return nil
	}
	return descResp.DescribeStacksOutput
}

func (cf *CFClient) validateTemplate(template []byte) *cloudformation.ValidateTemplateOutput {
	log.Println("Validating template.")
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(template)),
	}
	req := cf.Client.ValidateTemplateRequest(validateParams)
	resp, err := req.Send(context.Background())
	handle.Error(err)
	if resp == nil {
		handle.Fatal("the response was nil: ", errors.New("the response was nil"))
		return nil
	}
	return resp.ValidateTemplateOutput
}

// CDClient abstraction for cloudFormation client.
type CDClient struct {
	Client codedeployiface.ClientAPI
}

// IAMClient abstraction for cloudFormation client.
type IAMClient struct {
	Client iamiface.ClientAPI
}

// NotEmptyStack test structs which defines a non-empty stack.
var NotEmptyStack = cloudformation.DescribeStacksOutput{
	Stacks: []cloudformation.Stack{
		{
			StackName:   aws.String("TestStack"),
			StackStatus: cloudformation.StackStatusCreateComplete,
		},
	},
}
