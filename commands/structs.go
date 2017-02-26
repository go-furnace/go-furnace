package commands

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}

// NotEmptyStack test structs which defines a non-empty stack.
var NotEmptyStack = &cloudformation.DescribeStacksOutput{
	Stacks: []*cloudformation.Stack{
		&cloudformation.Stack{StackName: aws.String("TestStack")},
	},
}
