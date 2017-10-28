package awscommands

import (
	"errors"
	"log"
	"os"
	"reflect"
	"testing"

	config "github.com/Skarlso/go-furnace/config/common"
	"github.com/Skarlso/go-furnace/utils"
	commander "github.com/Yitsushi/go-commander"
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

func init() {
	utils.LogFatalf = log.Fatalf
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
	defer os.Clearenv()
	determineDeployment()
	if gitAccount != "test/account" {
		t.Fatalf("git account was not equal to test/account. Was: %s\n", gitAccount)
	}
	if gitRevision != "testrevision" {
		t.Fatalf("git revision was not equal to testrevision. Was: %s\n", gitRevision)
	}
}

func TestPushExecute(t *testing.T) {
	iamClient := new(IAMClient)
	iamClient.Client = &fakePushIAMClient{err: nil}
	cdClient := new(CDClient)
	cdClient.Client = &fakePushCDClient{err: nil, awsErr: nil}
	cfClient := new(CFClient)
	cfClient.Client = &fakePushCFClient{err: nil}
	opts := &commander.CommandHelper{}
	pushExecute(opts, cfClient, cdClient, iamClient)
}

func TestDetermineDeploymentS3(t *testing.T) {
	s3Deploy = true
	os.Setenv("FURNACE_S3BUCKET", "testBucket")
	os.Setenv("FURNACE_S3KEY", "testKey")
	defer os.Clearenv()
	determineDeployment()
	if s3Key != "testKey" {
		t.Fatalf("s3 key was not set. Was: %s\n", s3Key)
	}
	if codeDeployBucket != "testBucket" {
		t.Fatalf("s3 bucket was not set. Was: %s\n", codeDeployBucket)
	}
}

func TestDetermineDeploymentFailS3BucketNotSet(t *testing.T) {
	failed := false
	expectedMessage := "Please define FURNACE_S3BUCKET for the bucket to use."
	var message string
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	s3Deploy = true
	os.Setenv("FURNACE_S3KEY", "testKey")
	defer os.Clearenv()
	determineDeployment()
	if !failed {
		t.Error("should have failed execution")
	}
	if message != expectedMessage {
		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
	}
}

func TestDetermineDeploymentFailS3KeyNotSet(t *testing.T) {
	failed := false
	expectedMessage := "Please define FURNACE_S3KEY for the application to deploy."
	var message string
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	s3Deploy = true
	os.Setenv("FURNACE_S3BUCKET", "testBucket")
	defer os.Clearenv()
	determineDeployment()
	if !failed {
		t.Error("should have failed execution")
	}
	if message != expectedMessage {
		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
	}
}

func TestDetermineDeploymentFailGitAccountNotSet(t *testing.T) {
	s3Deploy = false
	failed := false
	expectedMessage := "Please define a git account and project to deploy from in the form of: account/project under FURNACE_GIT_ACCOUNT."
	var message string
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	os.Setenv("FURNACE_GIT_REVISION", "testrevision")
	defer os.Clearenv()
	determineDeployment()
	if !failed {
		t.Error("should have failed execution")
	}
	if message != expectedMessage {
		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
	}
}

func TestDetermineDeploymentFailGitRevisionNotSet(t *testing.T) {
	s3Deploy = false
	failed := false
	expectedMessage := "Please define the git commit hash to use for deploying under FURNACE_GIT_REVISION."
	var message string
	utils.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	os.Setenv("FURNACE_GIT_ACCOUNT", "test/account")
	defer os.Clearenv()
	determineDeployment()
	if !failed {
		t.Error("should have failed execution")
	}
	if message != expectedMessage {
		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
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

func TestCreateDeploymentGroupFailsOnNonAWSError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: errors.New("non aws error"), awsErr: nil}
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

func TestCreateApplicationFailsOnNonAWSError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: errors.New("non aws error"), awsErr: nil}
	err := createApplication("dummyApp", client)
	if err == nil {
		t.Fatal("error was not nil: ", err)
	}
}

func TestRevisionLocationS3(t *testing.T) {
	s3Deploy = true
	codeDeployBucket = "testBucket"
	s3Key = "testKey"
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
	defer os.Clearenv()
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

func TestPushCreate(t *testing.T) {
	wrapper := NewPush("furnace")
	if wrapper.Help.Arguments != "appName [-s3]" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"", "appName", "appName -s3", "-s3", "appName"}) ||
		wrapper.Help.LongDescription != `Push a version of the application to a stack` ||
		wrapper.Help.ShortDescription != "Push to stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
