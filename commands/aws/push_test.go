package awscommands

import (
	"errors"
	"log"
	"os"
	"reflect"
	"testing"

	config "github.com/Skarlso/go-furnace/config/common"
	commander "github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/codedeployiface"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/iamiface"
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
	config.LogFatalf = log.Fatalf
}

func (fiam *fakePushIAMClient) GetRoleRequest(*iam.GetRoleInput) iam.GetRoleRequest {
	return iam.GetRoleRequest{
		Request: &aws.Request{
			Data: &iam.GetRoleOutput{
				Role: &iam.Role{
					Arn: aws.String("CoolFakeRole"),
				},
			},
		},
	}
}

func (fc *fakePushCFClient) ListStackResourcesRequest(input *cloudformation.ListStackResourcesInput) cloudformation.ListStackResourcesRequest {
	var id *string
	if "NoASG" == *input.StackName {
		id = aws.String("NoASG")
	} else {
		id = aws.String("AWS::AutoScaling::AutoScalingGroup")
	}
	return cloudformation.ListStackResourcesRequest{
		Request: &aws.Request{
			Data: &cloudformation.ListStackResourcesOutput{
				StackResourceSummaries: []cloudformation.StackResourceSummary{
					{
						ResourceType:       id,
						PhysicalResourceId: aws.String("arn::whatever"),
					},
				},
			},
			Error: fc.err,
		},
	}
}

func (fd *fakePushCDClient) CreateDeploymentGroupRequest(input *codedeploy.CreateDeploymentGroupInput) codedeploy.CreateDeploymentGroupRequest {
	var err error
	if fd.awsErr != nil {
		err = fd.awsErr
	} else {
		err = fd.err
	}
	return codedeploy.CreateDeploymentGroupRequest{
		Request: &aws.Request{
			Data:  &codedeploy.CreateDeploymentGroupOutput{},
			Error: err,
		},
	}
}

func (fd *fakePushCDClient) CreateApplicationRequest(input *codedeploy.CreateApplicationInput) codedeploy.CreateApplicationRequest {
	var err error
	if fd.awsErr != nil {
		err = fd.awsErr
	} else {
		err = fd.err
	}
	return codedeploy.CreateApplicationRequest{
		Request: &aws.Request{
			Data:  &codedeploy.CreateApplicationOutput{},
			Error: err,
		},
	}
}

func (fd *fakePushCDClient) CreateDeploymentRequest(input *codedeploy.CreateDeploymentInput) codedeploy.CreateDeploymentRequest {
	return codedeploy.CreateDeploymentRequest{
		Request: &aws.Request{
			Data: &codedeploy.CreateDeploymentOutput{
				DeploymentId: aws.String("fakeID"),
			},
		},
	}
}

func (fd *fakePushCDClient) WaitUntilDeploymentSuccessful(input *codedeploy.GetDeploymentInput) error {
	return fd.err
}

func (fd *fakePushCDClient) GetDeploymentRequest(input *codedeploy.GetDeploymentInput) codedeploy.GetDeploymentRequest {
	return codedeploy.GetDeploymentRequest{
		Request: &aws.Request{
			Data: &codedeploy.GetDeploymentOutput{
				DeploymentInfo: &codedeploy.DeploymentInfo{
					Status: codedeploy.DeploymentStatusCreated,
				},
			},
		},
	}
}

func TestDetermineDeploymentGit(t *testing.T) {
	s3Deploy = false
	os.Setenv("AWS_FURNACE_GIT_ACCOUNT", "test/account")
	os.Setenv("AWS_FURNACE_GIT_REVISION", "testrevision")
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
	os.Setenv("AWS_FURNACE_GIT_ACCOUNT", "testAccount")
	os.Setenv("AWS_FURNACE_GIT_REVISION", "asdf12345")
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
	config.LogFatalf = func(s string, a ...interface{}) {
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
	config.LogFatalf = func(s string, a ...interface{}) {
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
	expectedMessage := "Please define a git account and project to deploy from in the form of: account/project under AWS_FURNACE_GIT_ACCOUNT."
	var message string
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	os.Setenv("AWS_FURNACE_GIT_REVISION", "testrevision")
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
	expectedMessage := "Please define the git commit hash to use for deploying under AWS_FURNACE_GIT_REVISION."
	var message string
	config.LogFatalf = func(s string, a ...interface{}) {
		failed = true
		message = s
	}
	os.Setenv("AWS_FURNACE_GIT_ACCOUNT", "test/account")
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
		t.Fatal("error was nil: ", err)
	}
}

func TestCreateDeploymentGroupFailsOnNonAWSError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: errors.New("non aws error"), awsErr: nil}
	err := createDeploymentGroup("dummyApp", "dummyRole", "dummyAsg", client)
	if err == nil {
		t.Fatal("error was nil: ", err)
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
		t.Fatal("error was nil: ", err)
	}
}

func TestCreateApplicationFailsOnNonAWSError(t *testing.T) {
	client := new(CDClient)
	client.Client = &fakePushCDClient{err: errors.New("non aws error"), awsErr: nil}
	err := createApplication("dummyApp", client)
	if err == nil {
		t.Fatal("error was nil: ", err)
	}
}

func TestRevisionLocationS3(t *testing.T) {
	s3Deploy = true
	codeDeployBucket = "testBucket"
	s3Key = "testKey"
	expected := &codedeploy.RevisionLocation{
		S3Location: &codedeploy.S3Location{
			Bucket:     aws.String("testBucket"),
			BundleType: codedeploy.BundleTypeZip,
			Key:        aws.String("testKey"),
			// Version:    aws.String("VersionId"), TODO: This needs improvement
		},
		RevisionType: codedeploy.RevisionLocationTypeS3,
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
		RevisionType: "GitHub",
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
