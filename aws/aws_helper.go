package aws

import (
  "fmt"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
)

// TestAWS testing AWS configuration.
func TestAWS() {
  sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
  fmt.Println(sess)
}
