package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/handle"
)

// Push command.
type Push struct {
	cdClient  *CDClient
	cfClient  *CFClient
	iamClient *IAMClient
}

var s3Deploy = false
var gitRevision string
var gitAccount string

// Execute defines what this command does.
func (p *Push) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	appName := awsconfig.Config.Aws.AppName
	stackname := awsconfig.Config.Main.Stackname
	s3Deploy = opts.Flags["s3"]
	asgName := getAutoScalingGroupKey(p.cfClient)
	role := getCodeDeployRoleARN(awsconfig.Config.Aws.CodeDeployRole, p.iamClient)
	err := createApplication(appName, p.cdClient)
	handle.Error(err)
	err = createDeploymentGroup(appName, role, asgName, stackname, p.cdClient)
	handle.Error(err)
	push(appName, asgName, stackname, p.cdClient)
}

func createDeploymentGroup(appName string, role string, asg string, stackname string, client *CDClient) error {
	var asgs []string
	if len(asg) > 0 {
		asgs = append(asgs, asg)
	}
	params := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(appName),
		DeploymentGroupName: aws.String(appName + "DeploymentGroup"),
		ServiceRoleArn:      aws.String(role),
		AutoScalingGroups:   asgs,
		LoadBalancerInfo: &codedeploy.LoadBalancerInfo{
			ElbInfoList: []codedeploy.ELBInfo{
				{
					Name: aws.String("ElasticLoadBalancer"),
				},
			},
		},
		Ec2TagFilters: []codedeploy.EC2TagFilter{
			{
				Key:   aws.String("fu_stage"),
				Type:  "KEY_AND_VALUE",
				Value: aws.String(stackname),
			},
		},
	}
	req := client.Client.CreateDeploymentGroupRequest(params)
	resp, err := req.Send(context.Background())
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
	resp, err := req.Send(context.Background())
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
	if s3Deploy {
		return &codedeploy.RevisionLocation{
			S3Location: &codedeploy.S3Location{
				Bucket:     aws.String(awsconfig.Config.Aws.CodeDeploy.S3Bucket),
				BundleType: "zip",
				Key:        aws.String(awsconfig.Config.Aws.CodeDeploy.S3Key),
				// Version:    aws.String("VersionId"), TODO: This needs improvement
			},
			RevisionType: "S3",
		}
	}
	return &codedeploy.RevisionLocation{
		GitHubLocation: &codedeploy.GitHubLocation{
			CommitId:   aws.String(awsconfig.Config.Aws.CodeDeploy.GitRevision),
			Repository: aws.String(awsconfig.Config.Aws.CodeDeploy.GitAccount),
		},
		RevisionType: "GitHub",
	}
}

func push(appName, asg, stackname string, client *CDClient) {
	log.Println("Stackname: ", stackname)
	var asgs []string
	if len(asg) > 0 {
		asgs = append(asgs, asg)
	}
	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(appName),
		IgnoreApplicationStopFailures: aws.Bool(true),
		DeploymentGroupName:           aws.String(appName + "DeploymentGroup"),
		Revision:                      revisionLocation(),
		TargetInstances: &codedeploy.TargetInstances{
			AutoScalingGroups: asgs,
			TagFilters: []codedeploy.EC2TagFilter{
				{
					Key:   aws.String("fu_stage"),
					Type:  "KEY_AND_VALUE",
					Value: aws.String(stackname),
				},
			},
		},
		UpdateOutdatedInstancesOnly: aws.Bool(false),
	}
	req := client.Client.CreateDeploymentRequest(params)
	resp, err := req.Send(context.Background())
	handle.Error(err)
	waitForFunctionWithStatusOutput("SUCCEEDED", config.WAITFREQUENCY, func() {
		err := client.Client.WaitUntilDeploymentSuccessful(context.Background(), &codedeploy.GetDeploymentInput{
			DeploymentId: resp.DeploymentId,
		})
		if err != nil {
			return
		}
	})
	fmt.Println()
	deploymentRequest := client.Client.GetDeploymentRequest(&codedeploy.GetDeploymentInput{
		DeploymentId: resp.DeploymentId,
	})
	deployment, err := deploymentRequest.Send(context.Background())
	handle.Error(err)
	log.Println("Deployment Status: ", deployment.DeploymentInfo.Status)
}

func getAutoScalingGroupKey(client *CFClient) string {
	params := &cloudformation.ListStackResourcesInput{
		StackName: aws.String(awsconfig.Config.Main.Stackname),
	}
	req := client.Client.ListStackResourcesRequest(params)
	resp, err := req.Send(context.Background())
	handle.Error(err)
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
	resp, err := req.Send(context.Background())
	handle.Error(err)
	return *resp.Role.Arn
}

// NewPush Creates a new Push command.
func NewPush(appName string) *commander.CommandWrapper {
	cfg, err := external.LoadDefaultAWSConfig()
	handle.Error(err)
	cd := codedeploy.New(cfg)
	cdClient := CDClient{cd}
	cf := cloudformation.New(cfg)
	cfClient := CFClient{cf}
	iam := iam.New(cfg)
	iamClient := IAMClient{iam}
	p := Push{
		cdClient:  &cdClient,
		cfClient:  &cfClient,
		iamClient: &iamClient,
	}
	return &commander.CommandWrapper{
		Handler: &p,
		Help: &commander.CommandDescriptor{
			Name:             "push",
			ShortDescription: "Push to stack",
			LongDescription:  `Push a version of the application to a stack`,
			Arguments:        "custom-config [-s3]",
			Examples:         []string{"", "custom-config", "custom-config -s3", "-s3"},
		},
	}
}
