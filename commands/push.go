package commands

import (
	"fmt"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codedeploy"
)

// Push command.
type Push struct {
}

// Execute defines what this command does.
func (c *Push) Execute(opts *commander.CommandHelper) {
	appName := opts.Arg(0)
	if len(appName) < 1 {
		appName = config.STACKNAME
	}
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cdClient := codedeploy.New(sess, nil)
	client := CDClient{cdClient}

	push(appName, &client)
}

func push(appName string, client *CDClient) {
	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(appName), // Required
		IgnoreApplicationStopFailures: aws.Bool(true),
		Revision: &codedeploy.RevisionLocation{
			GitHubLocation: &codedeploy.GitHubLocation{
				CommitId:   aws.String("CommitId"),
				Repository: aws.String("Repository"),
			},
		},
		TargetInstances: &codedeploy.TargetInstances{
			AutoScalingGroups: []*string{
				aws.String("AutoScalingGroup"),
			},
		},
		UpdateOutdatedInstancesOnly: aws.Bool(true),
	}
	resp, err := client.Client.CreateDeployment(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

// NewPush Creates a new Push command.
func NewPush(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Push{},
		Help: &commander.CommandDescriptor{
			Name:             "Push",
			ShortDescription: "Push to stack",
			LongDescription:  `Push a version of the application to a stack`,
			Arguments:        "name",
			Examples:         []string{"push", "push version"},
		},
	}
}
