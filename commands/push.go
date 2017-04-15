package commands

import (
	"fmt"
	"log"
	"os"

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

var s3Deploy = false
var codeDeployBucket string
var s3Key string

var gitRevision string
var gitAccount string

// Execute defines what this command does.
func (c *Push) Execute(opts *commander.CommandHelper) {
	sess := session.New(&aws.Config{Region: aws.String(config.REGION)})
	cd := codedeploy.New(sess, nil)
	cdClient := CDClient{cd}
	cf := cloudformation.New(sess, nil)
	cfClient := CFClient{cf}
	iam := iam.New(sess, nil)
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
	role := getCodeDeployRoleARN(config.CODEDEPLOYROLE, iamClient)
	err := createApplication(appName, cdClient)
	utils.CheckError(err)
	err = createDeploymentGroup(appName, role, asgName, cdClient)
	utils.CheckError(err)
	push(appName, asgName, cdClient)
}

func determineDeployment() {
	if s3Deploy {
		codeDeployBucket = os.Getenv("FURNACE_S3BUCKET")
		if len(codeDeployBucket) < 1 {
			utils.HandleFatal("Please define FURNACE_S3BUCKET for the bucket to use.", nil)
		}
		s3Key = os.Getenv("FURNACE_S3KEY")
		if len(s3Key) < 1 {
			utils.HandleFatal("Please define FURNACE_S3KEY for the application to deploy.", nil)
		}
		log.Println("S3 deployment will be used from bucket: ", codeDeployBucket)
	} else {
		gitAccount = os.Getenv("FURNACE_GIT_ACCOUNT")
		gitRevision = os.Getenv("FURNACE_GIT_REVISION")
		if len(gitAccount) < 1 {
			utils.HandleFatal("Please define a git account and project to deploy from in the form of: account/project under FURNACE_GIT_ACCOUNT.", nil)
		}
		if len(gitRevision) < 1 {
			utils.HandleFatal("Please define the git commit hash to use for deploying under FURNACE_GIT_REVISION.", nil)
		}
		log.Println("GitHub deployment will be used from account: ", gitAccount)
	}
}

func createDeploymentGroup(appName string, role string, asg string, client *CDClient) error {
	params := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(appName),
		DeploymentGroupName: aws.String(appName + "DeploymentGroup"),
		ServiceRoleArn:      aws.String(role),
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
	resp, err := client.Client.CreateApplication(params)
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
				BundleType: aws.String("zip"),
				Key:        aws.String(s3Key),
				// Version:    aws.String("VersionId"), TODO: This needs improvement
			},
			RevisionType: aws.String("S3"),
		}
	} else {
		rev = &codedeploy.RevisionLocation{
			GitHubLocation: &codedeploy.GitHubLocation{
				CommitId:   aws.String(gitRevision),
				Repository: aws.String(gitAccount),
			},
			RevisionType: aws.String("GitHub"),
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
			AutoScalingGroups: []*string{
				aws.String(asg),
			},
			TagFilters: []*codedeploy.EC2TagFilter{
				{
					Key:   aws.String("fu_stage"),
					Type:  aws.String("KEY_AND_VALUE"),
					Value: aws.String(config.STACKNAME),
				},
			},
		},
		UpdateOutdatedInstancesOnly: aws.Bool(false),
	}
	resp, err := client.Client.CreateDeployment(params)
	utils.CheckError(err)
	utils.WaitForFunctionWithStatusOutput("SUCCEEDED", config.WAITFREQUENCY, func() {
		client.Client.WaitUntilDeploymentSuccessful(&codedeploy.GetDeploymentInput{
			DeploymentId: resp.DeploymentId,
		})
	})
	fmt.Println()
	deployment, err := client.Client.GetDeployment(&codedeploy.GetDeploymentInput{
		DeploymentId: resp.DeploymentId,
	})
	utils.CheckError(err)
	log.Println("Deployment Status: ", *deployment.DeploymentInfo.Status)
}

func getAutoScalingGroupKey(client *CFClient) string {
	params := &cloudformation.ListStackResourcesInput{
		StackName: aws.String(config.STACKNAME),
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
			Arguments:        "appName [-s3]",
			Examples:         []string{"", "appName", "appName -s3", "-s3", "appName"},
		},
	}
}
