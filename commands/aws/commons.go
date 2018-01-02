package awscommands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	config "github.com/Skarlso/go-furnace/config/common"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
)

func gatherParameters(source *os.File, params *cloudformation.ValidateTemplateOutput) []cloudformation.Parameter {
	var stackParameters []cloudformation.Parameter
	defaultValue := color.New(color.FgHiBlack, color.Italic).SprintFunc()
	log.Println("Gathering parameters.")
	if params == nil {
		return stackParameters
	}
	for _, v := range params.Parameters {
		var param cloudformation.Parameter
		fmt.Printf("%s - '%s'(%s):", aws.StringValue(v.Description), keyName(aws.StringValue(v.ParameterKey)), defaultValue(aws.StringValue(v.DefaultValue)))
		text := readInputFrom(source)
		param.SetParameterKey(*v.ParameterKey)
		text = strings.Trim(text, "\n")
		if len(text) > 0 {
			param.SetParameterValue(*aws.String(text))
		} else {
			param.SetParameterValue(*v.DefaultValue)
		}
		stackParameters = append(stackParameters, param)
	}
	return stackParameters
}

func readInputFrom(source *os.File) string {
	reader := bufio.NewReader(source)
	text, _ := reader.ReadString('\n')
	return text
}

func (cf *CFClient) describeStacks(descStackInput *cloudformation.DescribeStacksInput) *cloudformation.DescribeStacksOutput {
	req := cf.Client.DescribeStacksRequest(descStackInput)
	descResp, err := req.Send()
	config.CheckError(err)
	return descResp
}

func (cf *CFClient) validateTemplate(template []byte) *cloudformation.ValidateTemplateOutput {
	log.Println("Validating template.")
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(template)),
	}
	req := cf.Client.ValidateTemplateRequest(validateParams)
	resp, err := req.Send()
	config.CheckError(err)
	return resp
}
