# Google Cloud

Google Cloud integration is a work in progress. Expect further update as it continue to be implemented.

Currently the supported and fully functional commands are:

```bash
❯ ./furnace-gcp help
delete custom-config           Delete a Google Deployment Manager
status [--config=configFile]   Get the status of an existing Deployment Management group.
create custom-config           Create a Google Deployment Manager
help [command]                 Display this help or a command specific help
```

## Authentication with Google

Please carefully read and follow the instruction outlined in this document: [Google Cloud Getting Started](https://cloud.google.com/sdk/#Quick_Start). It will describe how to download and install the SDK and initialize cloud to a Project Name and ID.

Take special attention to these documents:

[Initializing GCloud Tools](https://cloud.google.com/sdk/docs/initializing)
[Authorizing Tools](https://cloud.google.com/sdk/docs/authorizing)

Furnace uses a Google Key-File to authenticate with your Google Cloud Account and Project.
In the future, Furnace assumes these things are properly set up and in working order.

## Deployment Manager

Furnace uses Google Cloud's [Cloud Deployment Manager](https://cloud.google.com/deployment-manager/) service.
This service is similar to AWS' CloudFormation. It utilizes a YAML based configuration file and templates.
Templates use Python's [Jinja2](http://jinja.pocoo.org/) which is a fully featured template engine.

### Templates

You can find a LOT of good templates samples located here: [GloudPlatform Deployment Samples](https://github.com/GoogleCloudPlatform/deploymentmanager-samples). Furnace provides two examples. A simpler example can be seen in `./templates/google_template.yaml`. It will create a simple architecture with Load Balancing and Auto Scaling and deploy a Go Web App sample application located here: [Go Simple Wiki](https://github.com/Skarlso/furnace-google-cloud-app). It's Go's simple Wiki example app.

If deployed successfully, you should be able to access it like this:

![success](./img/working_go_app.png).

The second example can be located in `./templates/google_template.bookshelf.yaml`. This example deploys Google's sample Python App located here: [Python Getting Started](https://github.com/GoogleCloudPlatform/getting-started-python/tree/master/7-gce).

### Furnace Config

In order to tell furnace which template to use, simply create the following configuration file structure...

```
.
├── stacks
│   ├── google_template.yaml
│   ├── simple_template.jinja
│   └── gcp_furnace_config.yaml
└── .teststack.furnace
```

### Configuring a Deployment

*Note: The following section describes how to deploy the sample Go application.*

#### Setup

Project-ID should be set to your desired project name's ID with which to work with.

#### Update the template

Everything else, like region, is configured through the provided Google Templates. All attached `includes` and schema files are automatically added to the configuration. They should, however, live next to the template.

#### Startup Script

##### Store your startup script in a bucket

A startup script is what's used in order to bootstrap the instances. Furnace doesn't interpolate a script if it is attached, so rather use a bucket which contains the startup script and use `startup-script-url` template variable to define its location like this:

```yaml
      metadata:
        items:
          - key: startup-script-url
            value: gs://{{ properties["bucket"] }}/startup-script.sh
```

##### In-line with import

Right now, furnace doesn't provide an import from a schema file. A future version will have that luxury. The sample bookshelf template contains an example of that.

##### Raw in-line

You could always just in-line the script in the template directly.

## Creating a Deployment

After everything has been properly configure, execute:

```bash
./furnace-gcp create
```

This will display information like this:

```bash
~/golang/src/github.com/go-furnace/go-furnace extend_with_subcommand*
❯ ./furnace-gcp create
2017/11/03 07:14:47 Creating Deployment under project name: . testplatform-180405
2017/11/03 07:14:47 Deployment name is:  furnace-stack
2017/11/03 07:14:47 Found the following import files:  [{./simple_template.jinja simple_template.jinja}]
2017/11/03 07:14:47 Adding template name:  simple_template.jinja
2017/11/03 07:14:47 Looking for schema file for:  ./simple_template.jinja
2017/11/03 07:14:47 Schema to look for is:  /Users/hannibal/.config/go-furnace/simple_template.jinja.schema
[/] Waiting for state: DONE
```

## Deleting a Deployment

Once the stack is no longer needed, run the following command:

```bash
./furnace-gcp delete
```

Which will output this information:

```bash
~/golang/src/github.com/go-furnace/go-furnace extend_with_subcommand* 51s
❯ ./furnace-gcp delete
2017/11/03 07:17:38 Deleteing Deployment Under Project:  testplatform-180405
[-] Waiting for state: DONE
Stack terminated!
```

## Status of Deployment

Status can be retrieved using the following command:

```bash
./furnace-gcp status
```

This will output information about the deployment including the manifest file which includes all of the created resources with the deployment. This will look like the following output:


```bash
~/golang/src/github.com/go-furnace/go-furnace extend_with_subcommand* 1m 8s
❯ ./furnace-gcp status
2017/11/01 21:37:39 Status of Deployment under project name: . testplatform-180405
2017/11/01 21:37:39 Deployment name is:  furnace-stack
2017/11/01 21:37:41 Description:
2017/11/01 21:37:41 Name:  furnace-stack
2017/11/01 21:37:41 Labels:  []
2017/11/01 21:37:41 Selflink:  https://www.googleapis.com/deploymentmanager/v2/projects/testplatform-180405/global/deployments/furnace-stack
2017/11/01 21:37:41
Layout:
 resources:
- name: template
  properties:
    bucket: testplatform-180405.appspot.com
    machine-image: https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/family/debian-8
    machine-type: f1-micro
    max-instances: 1
    min-instances: 1
    scopes:
    - https://www.googleapis.com/auth/cloud-platform
    target-utilization: 0.6
    zone: europe-west3-a
  resources:
  - name: bookshelf-furnace-stack
    type: compute.v1.instanceTemplate
  - name: bookshelf-furnace-stack-frontend-group
    type: compute.v1.instanceGroupManager
  - name: bookshelf-furnace-stack-health-check
    type: compute.v1.httpHealthCheck
  - name: bookshelf-furnace-stack-frontend
    type: compute.v1.backendService
  - name: bookshelf-furnace-stack-frontend-map
    type: compute.v1.urlMap
  - name: bookshelf-furnace-stack-frontend-proxy
    type: compute.v1.targetHttpProxy
  - name: bookshelf-furnace-stack-frontend-http-rule
    type: compute.v1.globalForwardingRule
  - name: bookshelf-furnace-stack-autoscaler
    type: compute.v1.autoscaler
  - name: bookshelf-furnace-stack-allow-http
    type: compute.v1.firewall
  type: template.jinja
```