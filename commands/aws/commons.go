package awscommands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Skarlso/go-furnace/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
)

func gatherParameters(source *os.File, params *cloudformation.ValidateTemplateOutput) []*cloudformation.Parameter {
	var stackParameters []*cloudformation.Parameter
	defaultValue := color.New(color.FgHiBlack, color.Italic).SprintFunc()
	log.Println("Gathering parameters.")
	for _, v := range params.Parameters {
		var param cloudformation.Parameter
		fmt.Printf("%s - '%s'(%s):", *v.Description, keyName(*v.ParameterKey), defaultValue(*v.DefaultValue))
		text := readInputFrom(source)
		param.SetParameterKey(*v.ParameterKey)
		text = strings.Trim(text, "\n")
		if len(text) > 0 {
			param.SetParameterValue(*aws.String(text))
		} else {
			param.SetParameterValue(*v.DefaultValue)
		}
		stackParameters = append(stackParameters, &param)
	}
	return stackParameters
}

func readInputFrom(source *os.File) string {
	reader := bufio.NewReader(source)
	text, _ := reader.ReadString('\n')
	return text
}

func (cf *CFClient) describeStacks(descStackInput *cloudformation.DescribeStacksInput) *cloudformation.DescribeStacksOutput {
	descResp, err := cf.Client.DescribeStacks(descStackInput)
	utils.CheckError(err)
	return descResp
}

func (cf *CFClient) validateTemplate(template []byte) *cloudformation.ValidateTemplateOutput {
	log.Println("Validating template.")
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(template)),
	}
	resp, err := cf.Client.ValidateTemplate(validateParams)
	utils.CheckError(err)
	return resp
}
