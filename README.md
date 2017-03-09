# Furnace

![Logo](img/logo.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/Skarlso/go-furnace)](https://goreportcard.com/report/github.com/Skarlso/go-furnace) [![Build Status](https://travis-ci.org/Skarlso/go-furnace.svg?branch=master)](https://travis-ci.org/Skarlso/go-furnace)
[![Coverage Status](https://coveralls.io/repos/github/Skarlso/go-furnace/badge.svg?branch=master)](https://coveralls.io/github/Skarlso/go-furnace?branch=master)

## Intro

AWS Cloud Formation hosting with Go. This project utilizes the power of AWS CloudFormation and CodeDeploy in order to
simply deploy an application into a robust, self-healing, redundant environment. The environment is configurable through
the CloudFormation Template. A sample can be found in the `templates` folder.

The application to be deployed is handled via GitHub, or S3.

A sample application is provider under the `furnace-codedeploy-app` folder.

## AWS

### CloudFormation

[CloudFormation](https://aws.amazon.com/cloudformation/) as stated in the AWS documentation is an
> ...easy way to create and manage a collection of related AWS resources, provisioning and updating them in an orderly and predictable fashion.

Meaning, that via a template file it is possible to provide a description of the environment we would like to launch
are application into. How many server we would like to have? Load Balancing, and Auto Scaling setup. Own, isolated
network with VPCs. CloudFormation brings all these elements together into a bundler project called a `Stack`.
This stack can be created, updated, deleted and queried for various information.

This is what `Furnace` aims ti abstract in order to provide a very easy interface to work with complex architecture.

### CodeDeploy

[CodeDeploy](http://docs.aws.amazon.com/codedeploy/latest/userguide/welcome.html), as the documentation states
> ...coordinates application deployments to Amazon EC2 instances

In short, once the stack is up, we would like to deploy our application to the stack for usage. CodeDeploy takes care of that.
We don't have to scp something to our instances, we don't have to care if an instance goes away, or if we would like to have
a copy of that same instance. CodeDeploy can be integrated with various other services, so once we described how to deploy
our application, we never have to worry about it again. A simple `furnace push` will install our app to every instance that
the `Stack` manages.

Don't forget to install the CodeDeploy agent to your instances for the CodeDeploy to work. For this, see an example in the
provided template.

## Go

The decision to use [Go](https://golang.org/) for this project came very easy considering the many benefits Go provides when
handling APIs and async requests. Downloading massive files from S3 in threads, or starting many services at once is a breeze.
Go also provides a single binary which is easy to put on the execution path and use `Furnace` from any location.

Go has ample libraries which come to aid with AWS and their own Go SDK is pretty mature and stable.

## Usage

### Make

This project is using a `Makefile` for it's build processes. The following commands will create a binary and
run all tests:

```bash
make build test
```

`make install` will install the binary in your `$GOHOME\bin` path. Though the binary will be named `go-furnace`.

For other targets, please consult the Makefile.

### Configuration

Furnace has two stack related environment properties and a couple more which are shown later.

```bash
export FURNACE_REGION=eu-central-1
# If this is not defined, a default will be used which is FurnaceStack
export FURNACE_STACKNAME=FurnaceStack
```

Furnace also requires the CloudFormation template to be placed under `~/.config/go-furnace`.

CodeDeploy further requires an IAM policy on the current user in order to be able to handle ASG and deploying to the EC2 instances.
For this, a regular IAM role can be created from the AWS console. The name of the IAM profile can be configured later when pushing,
if that is not set, the default is used which is `CodeDeployServiceRole`. This default can be found under `config.CODEDEPLOYROLE`.

### Commands

Furnace provides the following commands (which you can check by running `./furnace`):

```bash
➜  go-furnace git:(master) ✗ ./furnace
create                    Create a stack
delete                    Delete a stack
status                    Status of a stack.
push appName [-s3]        Push to stack
delete-application name   Deletes an Application
help [command]            Display this help or a command specific help
```

Create and Delete will wait for the these actions to complete via a Waiter function. The waiters spinner type
can be set via the env property `FURNACE_SPINNER`. This is optional. The following spinners are available:

```go
// Spinners is a collection os spinner types
var Spinners = []string{`←↖↑↗→↘↓↙`,
	`▁▃▄▅▆▇█▇▆▅▄▃`,
	`┤┘┴└├┌┬┐`,
	`◰◳◲◱`,
	`◴◷◶◵`,
	`◐◓◑◒`,
	`⣾⣽⣻⢿⡿⣟⣯⣷`,
	`|/-\`}
```

The spinner defaults to `|/-\` which is # 7.

#### create

This will create the whole stack via the configuration provided under templates.

![Create](./img/create.png)

As you can see, furnace will ask for the parameters that reside in a template. If default is desired, simply
hit enter to continue using the default value.

#### delete

Deletes the whole stack complete with everything attached to the stack expect for the CodeDeploy application.

![Delete](./img/delete.png)

#### push

This is the command to get your application to be deployed onto all of your configured instances. This works via
two things. AutoScaling groups provided by the CloudFormation stack plus Tags that are put onto the instances called
`fu_stage`.

![Push](./img/push.png)

Push works with two revision locations.

##### GitHub

The default for a push is to locate a sample application on Github which will then be deployed.

For this, the following two options need to be defined:

```bash
export FURNACE_GIT_REVISION=b80ea5b9dfefcd21e27a3e0f149ec73519d5a6f1
export FURNACE_GIT_ACCOUNT=skarlso/furnace-codedeploy-app
```

##### S3

To use S3 for deployment, push needs an additional flag like this: `furnace push --s3`. This requires the following
two environment properties:

```bash
export FURNACE_S3KEY=app.zip
export FURNACE_S3BUCKET=furnace-codedeploy-bucket
```

Bucket is a unique bucket which is used to store a zipped version of the application. The key is the name of the object.
Access to the bucket needs to be defined in the CloudFormation template via an IAM Role. A sample is provided in the
template under the `templates` folder.

#### delete-application

Will delete the application and the deployment group completely.

#### successful push

If you are using the provided example and everything works, you should see the following output once you visit the
url provided by the load balancer.

![Success](./img/push_success.png)

## Plugins

### Experimental Plug-in System

To enable the plugin system, please set the environment property `FURNACE_ENABLE_PLUGIN_SYSTEM`.

To use the plugin system, please look at the example plugins in project `furnace-plugins`. At the moment, plugins are not receiving the
environment for further manipulations; however, this will be remedied.

A plugin is a standalone go project with main package, and a single entry point function called `RunPlugin`. If the plugin fails
to provide that function it will not be loaded. Plugins have to be placed under `~/.config/go-furnace/plugins`. Their extension decide
at what stage they will be loaded. Extensions should be one of the following: `pre_create, post_create, pre_delete, post_delete`. If not,
the plugin will simple be ignored. Well, technically it will be loaded, just not used.

To build the plugin run `go build -buildmode=plugin -o myplugin.pre_create myplugin.go`. Than copy the `.pre_create` to said folder. And
that's it, you should be all set.

Plugins are loaded as encountered, so if order of execution is important, pre_fix the file names with numbers.

**IMPORTANT**: Plugins are only supported on Linux right now. If you would like to play with it, I recommend using the official Docker golang
container which is easy to use. To link your project into the container, run it with the following command from the root of your project:

```bash
docker run --name furnace -it -v `pwd`:/go/src/github.com/Skarlso/go-furnace golang bash
```

Should any question arise, please don't hesitate to open an issue with the PreFix [Question].

## Configuration Management

Any kind of Configuration Management needs to be implemented by the application which is deployed.

That means that changes are applied to the `appspec.yml` file and the structure of the application itself.

For further examples checkout the AWS codedeploy example: [AwsLabs](https://github.com/awslabs/aws-codedeploy-samples).

## Testing

Testing the project for development is simply by executing `make test`.

## Contributions

Contributions are very welcomed, ideas, questions, remarks, please don't hesitate to submit a ticket. On what to do,
please take a look at the ROADMAP.md file.
