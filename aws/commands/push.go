package commands

import (
	"fmt"
	"log"
	"os"

	awsconfig "github.com/Skarlso/go-furnace/aws/config"
	config "github.com/Skarlso/go-furnace/config"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// Push command.
type Push struct {
}

var s3Deploy = false
var codeDeployBucket string
var s3Key string

var gitRevision string
var gitAccount string

// Execute defines what this command does.
func (c *Push) Execute(opts *commander.CommandHelper) {
	cfg, err := external.LoadDefaultAWSConfig()
	config.CheckError(err)
	cd := codedeploy.New(cfg)
	cdClient := CDClient{cd}
	cf := cloudformation.New(cfg)
	cfClient := CFClient{cf}
	iam := iam.New(cfg)
	iamClient := IAMClient{iam}
	pushExecute(opts, &cfClient, &cdClient, &iamClient)
}

func pushExecute(opts *commander.CommandHelper, cfClient *CFClient, cdClient *CDClient, iamClient *IAMClient) {
	appName := opts.Arg(1)
	if len(appName) < 1 {
		appName = config.STACKNAME
	}
	s3Deploy = opts.Flags["s3"]
	determineDeployment()
	asgName := getAutoScalingGroupKey(cfClient)
	role := getCodeDeployRoleARN(awsconfig.CODEDEPLOYROLE, iamClient)
	err := createApplication(appName, cdClient)
	config.CheckError(err)
	err = createDeploymentGroup(appName, role, asgName, cdClient)
	config.CheckError(err)
	push(appName, asgName, cdClient)
}

func determineDeployment() {
	if s3Deploy {
		codeDeployBucket = os.Getenv("FURNACE_S3BUCKET")
		if len(codeDeployBucket) < 1 {
			config.HandleFatal("Please define FURNACE_S3BUCKET for the bucket to use.", nil)
		}
		s3Key = os.Getenv("FURNACE_S3KEY")
		if len(s3Key) < 1 {
			config.HandleFatal("Please define FURNACE_S3KEY for the application to deploy.", nil)
		}
		log.Println("S3 deployment will be used from bucket: ", codeDeployBucket)
	} else {
		gitAccount = os.Getenv("AWS_FURNACE_GIT_ACCOUNT")
		gitRevision = os.Getenv("AWS_FURNACE_GIT_REVISION")
		if len(gitAccount) < 1 {
			config.HandleFatal("Please define a git account and project to deploy from in the form of: account/project under AWS_FURNACE_GIT_ACCOUNT.", nil)
		}
		if len(gitRevision) < 1 {
			config.HandleFatal("Please define the git commit hash to use for deploying under AWS_FURNACE_GIT_REVISION.", nil)
		}
		log.Println("GitHub deployment will be used from account: ", gitAccount)
	}
}

func createDeploymentGroup(appName string, role string, asg string, client *CDClient) error {
	params := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(appName),
		DeploymentGroupName: aws.String(appName + "DeploymentGroup"),
		ServiceRoleArn:      aws.String(role),
		AutoScalingGroups: []string{
			asg,
		},
		LoadBalancerInfo: &codedeploy.LoadBalancerInfo{
			ElbInfoList: []codedeploy.ELBInfo{
				{
					Name: aws.String("ElasticLoadBalancer"),
				},
			},
		},
	}
	req := client.Client.CreateDeploymentGroupRequest(params)
	resp, err := req.Send()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != codedeploy.ErrCodeDeploymentGroupAlreadyExistsException {
				log.Println(awsErr.Code())
				return err
			}
			log.Println("DeploymentGroup already exists. Nothing to do.")
			return nil
		}
		return err
	}
	log.Println(resp)
	return nil
}

func createApplication(appName string, client *CDClient) error {
	params := &codedeploy.CreateApplicationInput{
		ApplicationName: aws.String(appName),
	}
	req := client.Client.CreateApplicationRequest(params)
	resp, err := req.Send()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != codedeploy.ErrCodeApplicationAlreadyExistsException {
				log.Println(awsErr.Code())
				return err
			}
			log.Println("Application already exists. Nothing to do.")
			return nil
		}
		return err
	}
	log.Println(resp)
	return nil
}

func revisionLocation() *codedeploy.RevisionLocation {
	var rev *codedeploy.RevisionLocation
	if s3Deploy {
		rev = &codedeploy.RevisionLocation{
			S3Location: &codedeploy.S3Location{
				Bucket:     aws.String(codeDeployBucket),
				BundleType: "zip",
				Key:        aws.String(s3Key),
				// Version:    aws.String("VersionId"), TODO: This needs improvement
			},
			RevisionType: "S3",
		}
	} else {
		rev = &codedeploy.RevisionLocation{
			GitHubLocation: &codedeploy.GitHubLocation{
				CommitId:   aws.String(gitRevision),
				Repository: aws.String(gitAccount),
			},
			RevisionType: "GitHub",
		}
	}
	return rev
}

func push(appName string, asg string, client *CDClient) {
	log.Println("Stackname: ", config.STACKNAME)
	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(appName),
		IgnoreApplicationStopFailures: aws.Bool(true),
		DeploymentGroupName:           aws.String(appName + "DeploymentGroup"),
		Revision:                      revisionLocation(),
		TargetInstances: &codedeploy.TargetInstances{
			AutoScalingGroups: []string{
				asg,
			},
			TagFilters: []codedeploy.EC2TagFilter{
				{
					Key:   aws.String("fu_stage"),
					Type:  "KEY_AND_VALUE",
					Value: aws.String(config.STACKNAME),
				},
			},
		},
		UpdateOutdatedInstancesOnly: aws.Bool(false),
	}
	req := client.Client.CreateDeploymentRequest(params)
	resp, err := req.Send()
	config.CheckError(err)
	waitForFunctionWithStatusOutput("SUCCEEDED", config.WAITFREQUENCY, func() {
		client.Client.WaitUntilDeploymentSuccessful(&codedeploy.GetDeploymentInput{
			DeploymentId: resp.DeploymentId,
		})
	})
	fmt.Println()
	deploymentRequest := client.Client.GetDeploymentRequest(&codedeploy.GetDeploymentInput{
		DeploymentId: resp.DeploymentId,
	})
	deployment, err := deploymentRequest.Send()
	config.CheckError(err)
	log.Println("Deployment Status: ", deployment.DeploymentInfo.Status)
}

func getAutoScalingGroupKey(client *CFClient) string {
	params := &cloudformation.ListStackResourcesInput{
		StackName: aws.String(config.STACKNAME),
	}
	req := client.Client.ListStackResourcesRequest(params)
	resp, err := req.Send()
	config.CheckError(err)
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
	req := client.Client.GetRoleRequest(params)
	resp, err := req.Send()
	config.CheckError(err)
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
			Arguments:        "appName [-s3]",
			Examples:         []string{"", "appName", "appName -s3", "-s3", "appName"},
		},
	}
}
