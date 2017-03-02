package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
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
	createApplication(appName, &client)
	createDeploymentGroup(appName, &client)
	// push(appName, &client)
}

func createDeploymentGroup(appName string, client *CDClient) {
	// TODO: I have to get this from the CF stack.
	autoScalingGroupKeyName := "XXXX"
	params := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(appName),                                              // Required
		DeploymentGroupName: aws.String(appName + "DeploymentGroup"),                          // Required
		ServiceRoleArn:      aws.String("arn:aws:iam::xxxxxxxxxx:role/CodeDeployServiceRole"), // Required
		AutoScalingGroups: []*string{
			aws.String(autoScalingGroupKeyName),
		},
		LoadBalancerInfo: &codedeploy.LoadBalancerInfo{
			ElbInfoList: []*codedeploy.ELBInfo{
				{
					Name: aws.String("ElasticLoadBalancer"),
				},
			},
		},
	}
	resp, err := client.Client.CreateDeploymentGroup(params)
	utils.CheckError(err)
	log.Println(resp)
}

func createApplication(appName string, client *CDClient) {
	params := &codedeploy.CreateApplicationInput{
		ApplicationName: aws.String(appName), // Required
	}
	resp, err := client.Client.CreateApplication(params)
	utils.CheckError(err)
	log.Println(resp)
}

func push(appName string, client *CDClient) {
	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(appName), // Required
		IgnoreApplicationStopFailures: aws.Bool(true),
		DeploymentGroupName:           aws.String(appName + "DeploymentGroup"),
		Revision: &codedeploy.RevisionLocation{
			GitHubLocation: &codedeploy.GitHubLocation{
				CommitId:   aws.String("f1334d0ec8ea33abd2773bc8d6a475219bcf06f8"),
				Repository: aws.String("Skarlso/furnace-codedeploy-app"),
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
	utils.CheckError(err)
	log.Println(resp)
}

// NewPush Creates a new Push command.
func NewPush(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Push{},
		Help: &commander.CommandDescriptor{
			Name:             "push",
			ShortDescription: "Push to stack",
			LongDescription:  `Push a version of the application to a stack`,
			Arguments:        "name",
			Examples:         []string{"push", "push version"},
		},
	}
}
