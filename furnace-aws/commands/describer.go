package commands

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/go-furnace/go-furnace/handle"
)

func (cf *CFClient) describeStacks(descStackInput *cloudformation.DescribeStacksInput) *cloudformation.DescribeStacksOutput {
	req := cf.Client.DescribeStacksRequest(descStackInput)
	descResp, err := req.Send()
	handle.Error(err)
	return descResp
}
