package commands

import (
	"log"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Push command.
type Push struct {
}

// Execute defines what this command does.
func (c *Push) Execute(opts *commander.CommandHelper) {
	stackname := opts.Arg(0)
	if len(stackname) < 1 {
		stackname = config.STACKNAME
	}
	appName := opts.Arg(1)
	if len(appName) < 1 {
		appName = config.STACKNAME
	}
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cdClient := codedeploy.New(sess, nil)
	client := CDClient{cdClient}
	cf := cloudformation.New(sess, nil)
	cfClient := CFClient{cf}
	iam := iam.New(sess, nil)
	iamClient := IAMClient{iam}
	asgName := getAutoScalingGroupKey(stackname, &cfClient)
	role := getCodeDeployRoleARN(config.CODEDEPLOYROLE, &iamClient)
	createApplication(appName, &client)
	createDeploymentGroup(appName, role, asgName, &client)
	push(appName, stackname, asgName, &client)
}

func createDeploymentGroup(appName string, role string, asg string, client *CDClient) {
	params := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(appName),                     // Required
		DeploymentGroupName: aws.String(appName + "DeploymentGroup"), // Required
		ServiceRoleArn:      aws.String(role),                        // Required
		AutoScalingGroups: []*string{
			aws.String(asg),
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
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != codedeploy.ErrCodeDeploymentGroupAlreadyExistsException {
				log.Fatal(awsErr.Code())
			}
		} else {
			log.Fatal(err)
		}
	}
	log.Println(resp)
}

func createApplication(appName string, client *CDClient) {
	params := &codedeploy.CreateApplicationInput{
		ApplicationName: aws.String(appName), // Required
	}
	resp, err := client.Client.CreateApplication(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != codedeploy.ErrCodeApplicationAlreadyExistsException {
				log.Fatal(awsErr.Code())
			}
		} else {
			log.Fatal(err)
		}
	}
	log.Println(resp)
}

func push(appName string, stackname string, asg string, client *CDClient) {
	log.Println("Stackname: ", stackname)
	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(appName), // Required
		IgnoreApplicationStopFailures: aws.Bool(true),
		DeploymentGroupName:           aws.String(appName + "DeploymentGroup"),
		Revision: &codedeploy.RevisionLocation{
			GitHubLocation: &codedeploy.GitHubLocation{
				CommitId:   aws.String("f1334d0ec8ea33abd2773bc8d6a475219bcf06f8"),
				Repository: aws.String("Skarlso/furnace-codedeploy-app"),
			},
			RevisionType: aws.String("GitHub"),
		},
		TargetInstances: &codedeploy.TargetInstances{
			AutoScalingGroups: []*string{
				aws.String(asg),
			},
			TagFilters: []*codedeploy.EC2TagFilter{
				{ // Required
					Key:   aws.String("fu_stage"),
					Type:  aws.String("KEY_AND_VALUE"),
					Value: aws.String(stackname),
				},
			},
		},
		UpdateOutdatedInstancesOnly: aws.Bool(false),
	}
	resp, err := client.Client.CreateDeployment(params)
	utils.CheckError(err)
	log.Println(resp)
}

func getAutoScalingGroupKey(stackname string, client *CFClient) string {
	params := &cloudformation.ListStackResourcesInput{
		StackName: aws.String(stackname),
	}
	resp, err := client.Client.ListStackResources(params)
	utils.CheckError(err)
	for _, r := range resp.StackResourceSummaries {
		if *r.ResourceType == "AWS::AutoScaling::AutoScalingGroup" {
			return *r.PhysicalResourceId
		}
	}
	return ""
}

func getCodeDeployRoleARN(roleName string, client *IAMClient) string {
	params := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	resp, err := client.Client.GetRole(params)
	utils.CheckError(err)
	return *resp.Role.Arn
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
