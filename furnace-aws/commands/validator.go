package commands

import (
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/go-furnace/go-furnace/handle"
)

func (cf *CFClient) validateTemplate(template []byte) *cloudformation.ValidateTemplateOutput {
	log.Println("Validating template.")
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(template)),
	}
	req := cf.Client.ValidateTemplateRequest(validateParams)
	resp, err := req.Send()
	handle.Error(err)
	return resp
}
