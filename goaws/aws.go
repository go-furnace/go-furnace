package goaws

import (
	"log"
	"sync"
	"time"

	"github.com/Skarlso/go-furnace/errorhandler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// CreateCF Create a cloudformation stack.
func CreateCF(config []byte) {
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cfClient := cloudformation.New(sess, nil)
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(config)),
	}

	template, err := cfClient.ValidateTemplate(validateParams)
	errorhandler.CheckError(err)
	// if err != nil {
	// 	log.Fatal("Error occurred while validating cloudformation template. Please fix the following problem(s):", err)
	// }
	log.Println("The following template parameters will be asked for: ", template)
	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String("FurnaceStack"),
		TemplateBody: aws.String(string(config)),
	}
	resp, err := cfClient.CreateStack(stackInputParams)
	errorhandler.CheckError(err)
	log.Println("Create stack response: ", resp.GoString())
}

// WaitForFunctionWithStatusOutput waits for an ec2 function to complete its action.
func WaitForFunctionWithStatusOutput(status string, f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool)
	go func() {
		defer wg.Done()
		f()
		done <- true
	}()
	go func() {
		for {
			log.Println("Waiting for state: ", status)
			time.Sleep(1 * time.Second)
			select {
			case <-done:
				break
			default:
			}
		}
	}()

	wg.Wait()
}
