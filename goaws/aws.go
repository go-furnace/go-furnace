package goaws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Skarlso/go-furnace/errorhandler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var spinners = []string{`←↖↑↗→↘↓↙`, `▁▃▄▅▆▇█▇▆▅▄▃`, `┤┘┴└├┌┬┐`, `◰◳◲◱`, `◴◷◶◵`, `◐◓◑◒`, `⣾⣽⣻⢿⡿⣟⣯⣷`, `|/-\`}

// This will be configurable
var spinner = 6

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
	log.Println("The following template parameters will be asked for: ", template)
	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String("FurnaceStack"),
		TemplateBody: aws.String(string(config)),
	}
	resp, err := cfClient.CreateStack(stackInputParams)
	errorhandler.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String("FurnaceStack"),
	}
	log.Println("Create stack response: ", resp.GoString())
	WaitForFunctionWithStatusOutput("CREATE_COMPLETE", func() {
		cfClient.WaitUntilStackCreateComplete(describeStackInput)
	})
	descResp, err := cfClient.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String("FurnaceStack")})
	errorhandler.CheckError(err)
	log.Println("Stack state is: ", *descResp.Stacks[0].StackStatus)

}

// WaitForFunctionWithStatusOutput waits for a function to complete its action.
func WaitForFunctionWithStatusOutput(state string, f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool)
	go func() {
		defer wg.Done()
		f()
		done <- true
	}()
	go func() {
		counter := 0
		for {
			counter = (counter + 1) % len(spinners[spinner])
			fmt.Printf("\r\033[36m[%s]\033[m Waiting for stack to be in state: %s", string(spinners[spinner][counter]), state)
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
