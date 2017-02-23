package commands

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}
