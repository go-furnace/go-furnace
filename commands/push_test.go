package commands

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Skarlso/go-furnace/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codedeploy/codedeployiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type fakePushCFClient struct {
	cloudformationiface.CloudFormationAPI
	err error
}

type fakePushIAMClient struct {
	iamiface.IAMAPI
	err error
}

type fakePushCDClient struct {
	codedeployiface.CodeDeployAPI
	err    error
	awsErr awserr.Error
}

func (fiam *fakePushIAMClient) GetRole(*iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	return &iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("CoolFakeRole"),
		},
	}, fiam.err
}

func (fc *fakePushCFClient) ListStackResources(input *cloudformation.ListStackResourcesInput) (*cloudformation.ListStackResourcesOutput, error) {
	if "NoASG" == *input.StackName {
		return &cloudformation.ListStackResourcesOutput{
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{
					ResourceType:       aws.String("NoASG"),
					PhysicalResourceId: aws.String("arn::whatever"),
				},
			},
		}, fc.err
	}
	return &cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []*cloudformation.StackResourceSummary{
			{
				ResourceType:       aws.String("AWS::AutoScaling::AutoScalingGroup"),
				PhysicalResourceId: aws.String("arn::whatever"),
			},
		},
	}, fc.err
}

func (fd *fakePushCDClient) CreateDeploymentGroup(input *codedeploy.CreateDeploymentGroupInput) (*codedeploy.CreateDeploymentGroupOutput, error) {
	if fd.awsErr != nil {
		log.Println("Aws errorcode:", fd.awsErr.Code())
		return nil, fd.awsErr
	}
	return &codedeploy.CreateDeploymentGroupOutput{}, fd.err
}
func (fd *fakePushCDClient) CreateApplication(input *codedeploy.CreateApplicationInput) (*codedeploy.CreateApplicationOutput, error) {
	if fd.awsErr != nil {
		log.Println("Aws errorcode:", fd.awsErr.Code())
		return nil, fd.awsErr
	}
	return &codedeploy.CreateApplicationOutput{}, fd.err
}

func (fd *fakePushCDClient) CreateDeployment(input *codedeploy.CreateDeploymentInput) (*codedeploy.CreateDeploymentOutput, error) {
	return &codedeploy.CreateDeploymentOutput{DeploymentId: aws.String("fakeID")}, fd.err
}

func (fd *fakePushCDClient) WaitUntilDeploymentSuccessful(input *codedeploy.GetDeploymentInput) error {
	return fd.err
}

func (fd *fakePushCDClient) GetDeployment(input *codedeploy.GetDeploymentInput) (*codedeploy.GetDeploymentOutput, error) {
	return &codedeploy.GetDeploymentOutput{
		DeploymentInfo: &codedeploy.DeploymentInfo{
			Status: aws.String("I'm fine"),
		},
	}, fd.err
}

func TestDetermineDeploymentGit(t *testing.T) {
	s3Deploy = false
	os.Setenv("FURNACE_GIT_ACCOUNT", "test/account")
	os.Setenv("FURNACE_GIT_REVISION", "testrevision")
	determineDeployment()
	if gitAccount != "test/account" {
		t.Fatalf("git account was not equal to test/account. Was: %s\n", gitAccount)
	}
	if gitRevision != "testrevision" {
		t.Fatalf("git revision was not equal to testrevision. Was: %s\n", gitRevision)
	}
}

func TestDetermineDeploymentS3(t *testing.T) {
	s3Deploy = true
	os.Setenv("FURNACE_S3BUCKET", "testBucket")
	os.Setenv("FURNACE_S3KEY", "testKey")
	determineDeployment()
	if s3Key != "testKey" {
		t.Fatalf("s3 key was not set. Was: %s\n", s3Key)
	}
	if codeDeployBucket != "testBucket" {
		t.Fatalf("s3 bucket was not set. Was: %s\n", codeDeployBucket)
	}
}

func TestCreateDeploymentGroupSuccess(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: nil}
	err := createDeploymentGroup("dummyApp", "dummyRole", "dummyAsg", client)
	if err != nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestCreateDeploymentGroupAlreadyExists(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: awserr.New(codedeploy.ErrCodeDeploymentGroupAlreadyExistsException, "DeploymentGroup already exists", nil)}
	err := createDeploymentGroup("dummyApp", "dummyRole", "dummyAsg", client)
	if err != nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestCreateDeploymentGroupFailsOnDifferentError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: awserr.New(codedeploy.ErrCodeDeploymentGroupNameRequiredException, "Different error", nil)}
	err := createDeploymentGroup("dummyApp", "dummyRole", "dummyAsg", client)
	if err == nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestCreateApplicationSuccess(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: nil}
	err := createApplication("dummyApp", client)
	if err != nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestCreateApplicationAlreadyExists(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: awserr.New(codedeploy.ErrCodeApplicationAlreadyExistsException, "DeploymentGroup already exists", nil)}
	err := createApplication("dummyApp", client)
	if err != nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestCreateApplicationFailsOnDifferentError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: awserr.New(codedeploy.ErrCodeApplicationNameRequiredException, "Different error", nil)}
	err := createApplication("dummyApp", client)
	if err == nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestRevisionLocationS3(t *testing.T) {
	s3Deploy = true
	os.Setenv("FURNACE_S3BUCKET", "testBucket")
	os.Setenv("FURNACE_S3KEY", "testKey")
	expected := &codedeploy.RevisionLocation{
		S3Location: &codedeploy.S3Location{
			Bucket:     aws.String("testBucket"),
			BundleType: aws.String("zip"),
			Key:        aws.String("testKey"),
			// Version:    aws.String("VersionId"), TODO: This needs improvement
		},
		RevisionType: aws.String("S3"),
	}
	actual := revisionLocation()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual: %v did not equal expected: %v\n", actual, expected)
	}
}

func TestRevisionLocationGit(t *testing.T) {
	s3Deploy = false
	os.Setenv("FURNACE_S3BUCKET", "testBucket")
	os.Setenv("FURNACE_S3KEY", "testKey")
	expected := &codedeploy.RevisionLocation{
		GitHubLocation: &codedeploy.GitHubLocation{
			CommitId:   aws.String(gitRevision),
			Repository: aws.String(gitAccount),
		},
		RevisionType: aws.String("GitHub"),
	}
	actual := revisionLocation()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual: %v did not equal expected: %v\n", actual, expected)
	}
}

func TestPushBasic(t *testing.T) {
	config.WAITFREQUENCY = 0
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: nil, awsErr: nil}
	push("fakeApp", "fakeASG", client)
}

func TestGetAutoScalingGroupKeyIfASGExists(t *testing.T) {
	client := new(CFClient)
	client.Client = &fakePushCFClient{err: nil}
	asg := getAutoScalingGroupKey(client)
	if asg != "arn::whatever" {
		t.Fatal("arn did not match expected. Was: ", asg)
	}
}

func TestGetAutoScalingGroupKeyEmptyIfASGDoesNotExists(t *testing.T) {
	config.STACKNAME = "NoASG"
	client := new(CFClient)
	client.Client = &fakePushCFClient{err: nil}
	asg := getAutoScalingGroupKey(client)
	if asg != "" {
		t.Fatal("arn did not match expected. Was: ", asg)
	}
}

func TestGetCodeDeployRoleARN(t *testing.T) {
	client := new(IAMClient)
	client.Client = &fakePushIAMClient{err: nil}
	role := getCodeDeployRoleARN("fakeRole", client)
	if role != "CoolFakeRole" {
		t.Fatal("role did not match expected:", role)
	}
}
