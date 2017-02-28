package commands

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/codedeploy/codedeployiface"
)

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}

// CDClient abstraction for cloudFormation client.
type CDClient struct {
	Client codedeployiface.CodeDeployAPI
}

// NotEmptyStack test structs which defines a non-empty stack.
var NotEmptyStack = &cloudformation.DescribeStacksOutput{
	Stacks: []*cloudformation.Stack{
		{StackName: aws.String("TestStack")},
	},
}
