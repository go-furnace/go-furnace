package commands

import (
	"errors"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Skarlso/go-furnace/config"
	awsconfig "github.com/Skarlso/go-furnace/furnace-aws/config"
	"github.com/Skarlso/go-furnace/handle"
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
	handle.LogFatalf = log.Fatalf
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
	id := aws.String("AWS::AutoScaling::AutoScalingGroup")
	if "NoASG" == *input.StackName {
		id = aws.String("NoASG")
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
	err := fd.err
	if fd.awsErr != nil {
		err = fd.awsErr
	}
	return codedeploy.CreateDeploymentGroupRequest{
		Request: &aws.Request{
			Data:  &codedeploy.CreateDeploymentGroupOutput{},
			Error: err,
		},
	}
}

func (fd *fakePushCDClient) CreateApplicationRequest(input *codedeploy.CreateApplicationInput) codedeploy.CreateApplicationRequest {
	err := fd.err
	if fd.awsErr != nil {
		err = fd.awsErr
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

func TestPushExecute(t *testing.T) {
	awsconfig.Config = awsconfig.Configuration{}
	awsconfig.Config.Aws.CodeDeploy.GitAccount = "test/account"
	awsconfig.Config.Aws.CodeDeploy.GitRevision = "testrevision"
	iamClient := new(IAMClient)
	iamClient.Client = &fakePushIAMClient{err: nil}
	cdClient := new(CDClient)
	cdClient.Client = &fakePushCDClient{err: nil, awsErr: nil}
	cfClient := new(CFClient)
	cfClient.Client = &fakePushCFClient{err: nil}
	opts := &commander.CommandHelper{}
	pushExecute(opts, cfClient, cdClient, iamClient)
}

func TestPushExecuteWithStackConfig(t *testing.T) {
	awsconfig.Config = awsconfig.Configuration{}
	awsconfig.Config.Aws.CodeDeploy.GitAccount = "test/account"
	awsconfig.Config.Aws.CodeDeploy.GitRevision = "testrevision"
	iamClient := new(IAMClient)
	iamClient.Client = &fakePushIAMClient{err: nil}
	cdClient := new(CDClient)
	cdClient.Client = &fakePushCDClient{err: nil, awsErr: nil}
	cfClient := new(CFClient)
	cfClient.Client = &fakePushCFClient{err: nil}
	opts := &commander.CommandHelper{}
	opts.Args = append(opts.Args, "teststack")
	pushExecute(opts, cfClient, cdClient, iamClient)
	if awsconfig.Config.Main.Stackname != "MyStack" {
		t.Fatal("test did not load the file requested.")
	}
}

func TestPushExecuteWithStackConfigNotFound(t *testing.T) {
	failed := false
	handle.LogFatalf = func(s string, a ...interface{}) {
		failed = true
	}
	awsconfig.Config = awsconfig.Configuration{}
	awsconfig.Config.Aws.CodeDeploy.GitAccount = "test/account"
	awsconfig.Config.Aws.CodeDeploy.GitRevision = "testrevision"
	iamClient := new(IAMClient)
	iamClient.Client = &fakePushIAMClient{err: nil}
	cdClient := new(CDClient)
	cdClient.Client = &fakePushCDClient{err: nil, awsErr: nil}
	cfClient := new(CFClient)
	cfClient.Client = &fakePushCFClient{err: nil}
	opts := &commander.CommandHelper{}
	opts.Args = append(opts.Args, "notfound")
	pushExecute(opts, cfClient, cdClient, iamClient)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
	}
}

// func TestDetermineDeploymentFailS3BucketNotSet(t *testing.T) {
// 	failed := false
// 	expectedMessage := "Please define S3BUCKET for the bucket to use."
// 	var message string
// 	handle.LogFatalf = func(s string, a ...interface{}) {
// 		failed = true
// 		message = s
// 	}
// 	s3Deploy = true
// 	awsconfig.Config = awsconfig.Configuration{}
// 	awsconfig.Config.Aws.CodeDeploy.S3Bucket = ""
// 	awsconfig.Config.Aws.CodeDeploy.S3Key = "key"
// 	defer os.Clearenv()
// 	if !failed {
// 		t.Error("should have failed execution")
// 	}
// 	if message != expectedMessage {
// 		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
// 	}
// }

// func TestDetermineDeploymentFailS3KeyNotSet(t *testing.T) {
// 	failed := false
// 	expectedMessage := "Please define S3KEY for the application to deploy."
// 	var message string
// 	handle.LogFatalf = func(s string, a ...interface{}) {
// 		failed = true
// 		message = s
// 	}
// 	s3Deploy = true
// 	awsconfig.Config = awsconfig.Configuration{}
// 	awsconfig.Config.Aws.CodeDeploy.S3Bucket = "testbucket"
// 	awsconfig.Config.Aws.CodeDeploy.S3Key = ""
// 	if !failed {
// 		t.Error("should have failed execution")
// 	}
// 	if message != expectedMessage {
// 		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
// 	}
// }

// func TestDetermineDeploymentFailGitAccountNotSet(t *testing.T) {
// 	s3Deploy = false
// 	failed := false
// 	expectedMessage := "Please define a git account and project to deploy from in the form of: account/project under GIT_ACCOUNT."
// 	var message string
// 	handle.LogFatalf = func(s string, a ...interface{}) {
// 		failed = true
// 		message = s
// 	}
// 	awsconfig.Config = awsconfig.Configuration{}
// 	awsconfig.Config.Aws.CodeDeploy.GitRevision = "revision"
// 	awsconfig.Config.Aws.CodeDeploy.GitAccount = ""
// 	defer os.Clearenv()
// 	if !failed {
// 		t.Error("should have failed execution")
// 	}
// 	if message != expectedMessage {
// 		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
// 	}
// }

// func TestDetermineDeploymentFailGitRevisionNotSet(t *testing.T) {
// 	s3Deploy = false
// 	failed := false
// 	expectedMessage := "Please define the git commit hash to use for deploying under GIT_REVISION."
// 	var message string
// 	handle.LogFatalf = func(s string, a ...interface{}) {
// 		failed = true
// 		message = s
// 	}
// 	awsconfig.Config = awsconfig.Configuration{}
// 	awsconfig.Config.Aws.CodeDeploy.GitAccount = "account"
// 	awsconfig.Config.Aws.CodeDeploy.GitRevision = ""
// 	if !failed {
// 		t.Error("should have failed execution")
// 	}
// 	if message != expectedMessage {
// 		t.Errorf("expected message %s did not equal actual %s", expectedMessage, message)
// 	}
// }

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

	expected := &codedeploy.RevisionLocation{
		S3Location: &codedeploy.S3Location{
			Bucket:     aws.String("testBucket"),
			BundleType: codedeploy.BundleTypeZip,
			Key:        aws.String("testKey"),
			// Version:    aws.String("VersionId"), TODO: This needs improvement
		},
		RevisionType: codedeploy.RevisionLocationTypeS3,
	}
	awsconfig.Config.Aws.CodeDeploy.S3Bucket = "testBucket"
	awsconfig.Config.Aws.CodeDeploy.S3Key = "testKey"
	actual := revisionLocation()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual: %v did not equal expected: %v\n", actual, expected)
	}
}

func TestRevisionLocationGit(t *testing.T) {
	s3Deploy = false
	awsconfig.Config = awsconfig.Configuration{}
	awsconfig.Config.Aws.CodeDeploy.S3Bucket = "testbucket"
	awsconfig.Config.Aws.CodeDeploy.S3Key = "testkey"
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
	client := new(CFClient)
	client.Client = &fakePushCFClient{err: nil}
	awsconfig.Config.Main.Stackname = "NoASG"
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
	if wrapper.Help.Arguments != "custom-config [-s3]" ||
		!reflect.DeepEqual(wrapper.Help.Examples, []string{"", "custom-config", "custom-config -s3", "-s3"}) ||
		wrapper.Help.LongDescription != `Push a version of the application to a stack` ||
		wrapper.Help.ShortDescription != "Push to stack" {
		t.Log(wrapper.Help.LongDescription)
		t.Log(wrapper.Help.ShortDescription)
		t.Log(wrapper.Help.Examples)
		t.Fatal("wrapper did not match with given params")
	}
}
