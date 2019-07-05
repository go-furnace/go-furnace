package commands

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"

	commander "github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/codedeployiface"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/iamiface"
	"github.com/go-furnace/go-furnace/config"
	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/handle"
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
			HTTPRequest: new(http.Request),
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
			Error:       fc.err,
			HTTPRequest: new(http.Request),
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
			Data:        &codedeploy.CreateDeploymentGroupOutput{},
			Error:       err,
			HTTPRequest: new(http.Request),
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
			Data:        &codedeploy.CreateApplicationOutput{},
			Error:       err,
			HTTPRequest: new(http.Request),
		},
	}
}

func (fd *fakePushCDClient) CreateDeploymentRequest(input *codedeploy.CreateDeploymentInput) codedeploy.CreateDeploymentRequest {
	return codedeploy.CreateDeploymentRequest{
		Request: &aws.Request{
			Data: &codedeploy.CreateDeploymentOutput{
				DeploymentId: aws.String("fakeID"),
			},
			HTTPRequest: new(http.Request),
		},
	}
}

func (fd *fakePushCDClient) WaitUntilDeploymentSuccessful(ctx context.Context, input *codedeploy.GetDeploymentInput, opts ...aws.WaiterOption) error {
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
			HTTPRequest: new(http.Request),
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
	p := Push{
		cfClient: cfClient,
		cdClient: cdClient,
		iamClient: iamClient,
	}
	p.Execute(opts)
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
	p := Push{
		cfClient: cfClient,
		cdClient: cdClient,
		iamClient: iamClient,
	}
	p.Execute(opts)
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
	p := Push{
		cfClient: cfClient,
		cdClient: cdClient,
		iamClient: iamClient,
	}
	p.Execute(opts)
	if !failed {
		t.Error("Expected outcome to fail. Did not fail.")
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
