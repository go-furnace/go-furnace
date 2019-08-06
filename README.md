# Furnace

![Logo](img/logo.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/go-furnace/go-furnace)](https://goreportcard.com/report/github.com/go-furnace/go-furnace) [![Build Status](https://travis-ci.org/go-furnace/go-furnace.svg?branch=master)](https://travis-ci.org/go-furnace/go-furnace)
[![Coverage Status](https://coveralls.io/repos/github/go-furnace/go-furnace/badge.svg?branch=master)](https://coveralls.io/github/go-furnace/go-furnace?branch=master)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go/)

## Intro

## Brief Explanation

Here is a very short description of what Furnace does in a handy IKEA manual format.

![Furnace1](img/ikea-furnace-1.png)
![Furnace2](img/ikea-furnace-2.png)

## In More Depth

AWS Cloud Formation, Google Cloud Platform, or DigitalOcean hosting with Go. This project utilizes the power of
AWS CloudFormation and CodeDeploy, or DeploymentManager and Git support in GCP in order to simply deploy an application
into a robust, self-healing, redundant environment. The environment is configurable through the CloudFormation Template or
GCPs jinja files. A sample can be found in the `templates` folder.

The application to be deployed is handled via GitHub, or S3.

A sample application is provider under the `furnace-codedeploy-app` folder.

## Furnace vs Terraform

Furnace does not try to compete with Terraform. It aims for a different market. The main differences between Terraform and Furnace
are the following:

### Binary size

On the rare occasions when disk space matters, Furnace provides individual binaries. The size of the aws binary ATM is 15MB.
Terraform is at 100MB.

### Configuration

Configuration for Terraform can be pretty huge. In contrast Furnace's configuration is lightweight because frankly, it doesn't
have much. All the configuration you will have to write will be for CloudFormation and GCP. However complicated they can be is out
of Furnace's reach.

### Vendor Lock

True that Terraform provides provider agnostic settings and behavior. But, you'll be vendor locked to Terraform. Moving away from
the massive configuration that the user has to build up to use it can never be moved away from again. Or only through a lot of
work.

In contrast, Furance is a light wrapper around services that provide what Terraform is providing per provider. What does this
mean? It means Furnace is using CloudFormation for AWS and DeploymentManager for GCP which are services built by AWS and GCP. Not
by Furnace. If you don't want to use Furnace any longer, you'd still have your deployment configuration which works just fine
without it. Resources are all grouped together. Deleting them is as simple as calling an end-point, clicking a button or hitting
enter.

## Installing Binaries

### Go Install

To install all generated binaries at once, run:

```bash
# Download / Clone the latest version
# cd into go-furnace
make install-all
```

This will install all dependencies and both binaries to `$GOPATH/bin` folder.

### Make commands

You can also build the commands which will be output into the `cmd` sub-folder.

```bash
# Simply run make from the root folder
make
```

### Building for different environment

Convenient targets are provided for linux and windows binaries.

```bash
make linux
make windows
```

These are only available from the package folders respectively.

### Clean

In case `make install` is used, a clean-up target is also provided.

```bash
make clean-all
```

## AWS

- [AWS](./furnace-aws/README.md)

## Google Cloud

- [Google Cloud](./furnace-gcp/README.md)

## DigitalOcean

- [DigitalOcean](./furnace-do/README.md)

## Plugins

A highly customizable plugin system is provided for Furnace via [HashiCorp's Go-Plugins](https://github.com/hashicorp/go-plugin).

Writing a plugin is as easy as implementing an interface. Furnace uses GRPC to talk to the plugins locally. The interface to
implement is provided by a proto file located here: [Protocol Description](https://github.com/go-furnace/proto).

A single configuration value is provided for plugins in the yaml file which is the location of plugins:

```yaml
  plugins:
    plugin_path: "./plugins"
```

If this is not provided, the default value is `./plugins` which is next to the binary.

Plugins are available for the following events:

* Pre creating a stack (stackname parameter is provided)
  These plugins have the chance to stop the process before it starts. Here the user would typically try and do a preliminary check
  like permissions or resources are available. If not, abort the creation process before it begins.

* Post creating a stack (stackname parameter is provided)
  This is typically a place where a post notification could be executed, like a slack notifier that a stack's creation is done.
  Or an application health-check which looks up the deployed URL parameter and checks if the application is responding.

* Pre deleting a stack (stackname parameter is provided)
  These plugins also have the option to abort a delete before it begins. A typical use-case would be to check if the resources
  associated to the stack are still being used or not.

* Post deleting a stack (stackname parameter is provided)
  This is a place to send out a notification that a stack has been successfully or unsuccessfully deleted.
  Or another application could be to see if all the resources where cleaned up properly. Or to perform any more cleanup
  which the CloudFormation could not do.

The following repository contains the SDK that the plugins provide for a Go based plugin system:

[SDK for Go based plugins](https://github.com/go-furnace/sdk).

### Multiple languages

Since it's GRPC the language in which the plugin is provided is whatever the plugin's writer chooses and is supported by Furnace.

Currently three main languages are supported to write plugins in:

* Python
* Ruby
* Go

### Slack Plugin in Go

```go
package main

import (
	"log"

	fplugins "github.com/go-furnace/go-furnace/furnace-aws/plugins"
	"github.com/go-furnace/sdk"
	"github.com/hashicorp/go-plugin"
)

// SlackPreCreate is an actual implementation of the furnace PreCreate plugin
// interface.
type SlackPreCreate struct{}

// Execute is the entry point to this plugin.
func (SlackPreCreate) Execute(stackname string) bool {
	api := slack.New("YOUR_TOKEN_HERE")
	params := slack.PostMessageParameters{}
	channelID, timestamp, err := api.PostMessage("#general", fmt.Sprintf("Stack with name '%s' is Done.", stackname), params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return true
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: fplugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			"slack-furnace-precreate": &sdk.PreCreateGRPCPlugin{Impl: &SlackPreCreate{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

### Sample plugin in Python

For this to work the author has to implement the proto file. A sample repository can be found here:
[Example for a Python Plugin](https://github.com/go-furnace/python-plugin).

For brevity here is the full Python source:

```python
from concurrent import futures
import sys
import time

import grpc

import furnace_pb2
import furnace_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

class PreCreatePluginServicer(furnace_pb2_grpc.PreCreateServicer):
    """Implementation of PreCreatePlugin service."""

    def Execute(self, request, context):
        result = furnace_pb2.Proceed()
        result.failed = True

        return result

def serve():
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    furnace_pb2_grpc.add_PreCreateServicer_to_server(PreCreatePluginServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port('127.0.0.1:1234')
    server.start()

    # Output information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
```

The serve method here is a `go-plugin` requirement. To read up on it, please check-out go-plugin by HashiCorp.

### Usage

After a plugin has been written simply build ( in case of Go ) or copy ( in case of Python ) it to the right location.

Furnace autodiscovers these files based on their name and loads them in order. Once that happens it will run them
together at the correct event.

The following filenames should be used for the following events:

* PreCreate: `*-furnace-precreate*`
* PostCreate: `*-furnace-postcreate*`
* PreDelete: `*-furnace-predelete*`
* PostDelete: `*-furnace-postdelete*`

## Separate binaries

In order to try and minimize the binary size of furnace, it has separate binaries for each service it provides.

You can find `furnace-aws` under `aws` and `furnace-gcp` under `gcp`. This way, if you plan on using only aws you don't need to worry about dependencies for Google, and vica-versa.

## Contributions

Contributions are very welcomed, ideas, questions, remarks, please don't hesitate to submit a ticket. On what to do,
please take a look at the [ROADMAP.md](./ROADMAP.md) file or under the Issues tab.

## Pre-Binaries

Are now available under release artifacts and are automatically built by CircleCI whenever a new tag is created.
