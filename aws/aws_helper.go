package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateEC2 testing AWS configuration.
func CreateEC2() {
	log.Println("Creating ec2 session.")
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	ec2Client := ec2.New(sess, nil)
	// ec2Client.RunInstances(nil)
	resp, err := ec2Client.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}
	log.Println(resp)
}
