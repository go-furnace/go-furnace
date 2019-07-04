package commands

import (
	"os"
	"reflect"
	"testing"

	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	dm "google.golang.org/api/deploymentmanager/v2"
)

func TestExecute(t *testing.T) {
	expectedDeployments := &dm.Deployment{
		Description: "",
		Fingerprint: "",
		Id:          0,
		InsertTime:  "",
		Labels:      nil,
		Manifest:    "",
		Name:        "teststack",
		Operation:   nil,
		SelfLink:    "",
		Target: &dm.TargetConfiguration{
			Config: &dm.ConfigFile{
				Content:         "# Copyright 2015 Google Inc. All rights reserved.\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n\n# [START all]\nimports:\n- name: simple_template.jinja\n  path: ./simple_template.jinja\n\nresources:\n- name: simple_template\n  type: simple_template.jinja\n  properties:\n    zone: europe-west3-a\n    machine-type: f1-micro\n    machine-image: https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/family/debian-8\n    min-instances: 1\n    max-instances: 1\n    target-utilization: 0.6\n    bucket: test-project-bucket\n    scopes:\n    - https://www.googleapis.com/auth/cloud-platform\n\n# [END all]\n",
				ForceSendFields: nil,
				NullFields:      nil,
			},
			Imports: []*dm.ImportFile{
				{
					Content:         "{#\nCopyright 2016 Google Inc. All rights reserved.\nLicensed under the Apache License, Version 2.0 (the \"License\");\nyou may not use this file except in compliance with the License.\nYou may obtain a copy of the License at\n    http://www.apache.org/licenses/LICENSE-2.0\nUnless required by applicable law or agreed to in writing, software\ndistributed under the License is distributed on an \"AS IS\" BASIS,\nWITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\nSee the License for the specific language governing permissions and\nlimitations under the License.\n#}\n\n{# [START all] #}\n\n{# [START env] #}\n{% set NAME = \"go-\" + env[\"deployment\"] %}\n{% set SERVICE = \"go-\" + env[\"deployment\"] + \"-frontend\" %}\n{# [END env] #}\n\n#\n# Instance group setup\n#\n\n# First we have to create an instance template.\n# This template will be used by the instance group\n# to create new instances.\nresources:\n- name : {{ NAME }}\n  type: compute.v1.instanceTemplate\n  properties:\n    properties:\n      tags:\n        items:\n          - http-server\n      disks:\n        - boot: True\n          type: PERSISTENT\n          initializeParams:\n            sourceImage: {{ properties['machine-image'] }}\n            diskSizeGb: 10\n            diskType: pd-ssd\n      machineType: {{ properties['machine-type'] }}\n      serviceAccounts:\n          - email: default\n            scopes: {{ properties['scopes'] }}\n      metadata:\n        items:\n          - key: startup-script-url\n            value: gs://{{ properties[\"bucket\"] }}/startup-script.sh\n      networkInterfaces:\n          - network: global/networks/default\n            accessConfigs:\n              - type: ONE_TO_ONE_NAT\n                name: External NAT\n\n# Creates the managed instance group. This is responsible for creating\n# new instances using the instance template, as well as providing a named\n# port the backend service can target\n- name: {{ NAME }}-frontend-group\n  type: compute.v1.instanceGroupManager\n  properties:\n    instanceTemplate: $(ref.{{  NAME  }}.selfLink)\n    baseInstanceName: frontend-group\n    targetSize: 3\n    zone: {{ properties['zone'] }}\n    namedPorts:\n      - name: http\n        port: 8080\n\n\n\n# Load Balancer Setup\n#\n\n# A complete HTTP load balancer is structured as follows:\n#\n# 1) A global forwarding rule directs incoming requests to a target HTTP proxy.\n# 2) The target HTTP proxy checks each request against a URL map to determine the\n#    appropriate backend service for the request.\n# 3) The backend service directs each request to an appropriate backend based on\n#    serving capacity, zone, and instance health of its attached backends. The\n#    health of each backend instance is verified using either a health check.\n#\n# We'll create these resources in reverse order:\n# service, health check, backend service, url map, proxy.\n\n# Create a health check\n# The load balancer will use this check to keep track of which instances to send traffic to.\n# Note that health checks will not cause the load balancer to shutdown any instances.\n- name: {{ NAME }}-health-check\n  type: compute.v1.httpHealthCheck\n  properties:\n    requestPath: /_ah/health\n    port: 8080\n\n# Create a backend service, associate it with the health check and instance group.\n# The backend service serves as a target for load balancing.\n- name: {{ SERVICE }}\n  type: compute.v1.backendService\n  properties:\n    healthChecks:\n      - $(ref.{{ NAME }}-health-check.selfLink)\n    portName: http\n    backends:\n{# [START reference] #}\n      - group: $(ref.{{ NAME }}-frontend-group.instanceGroup)\n        zone: {{ properties['zone'] }}\n{# [END reference] #}\n\n# Create a URL map and web Proxy. The URL map will send all requests to the\n# backend service defined above.\n- name: {{ SERVICE }}-map\n  type: compute.v1.urlMap\n  properties:\n    defaultService: $(ref.{{ SERVICE }}.selfLink)\n\n# This is the actual proxy which uses the URL map to route traffic\n# to the backend service\n- name: {{ SERVICE }}-proxy\n  type: compute.v1.targetHttpProxy\n  properties:\n    urlMap: $(ref.{{ SERVICE }}-map.selfLink)\n\n# This is the global forwarding rule which creates an external IP to\n# target the http poxy\n- name: {{ SERVICE }}-http-rule\n  type: compute.v1.globalForwardingRule\n  properties:\n    target: $(ref.{{ SERVICE }}-proxy.selfLink)\n    portRange: 80\n\n# Creates an autoscaler resource (note that when using the gcloud CLI,\n# autoscaling is set as a configuration of the managed instance group\n# but autoscaler is a resource so in deployment manager we explicitly\n# define it\n- name: {{ NAME }}-autoscaler\n  type: compute.v1.autoscaler\n  properties:\n    zone: {{ properties['zone'] }}\n    target: $(ref.{{ NAME }}-frontend-group.selfLink)\n    autoscalingPolicy:\n{# [START properties] #}\n      minNumReplicas: {{ properties['min-instances'] }}\n      maxNumReplicas: {{ properties['max-instances'] }}\n      loadBalancingUtilization:\n        utilizationTarget: {{ properties['target-utilization'] }}\n{# [END properties] #}\n\n# Firewall rule that allows traffic to GCE instances with the\n# http server tag we created\n- name: {{ NAME }}-allow-http\n  type: compute.v1.firewall\n  properties:\n    allowed:\n      - IPProtocol: tcp\n        ports:\n          - 8080\n    sourceRanges:\n      - 0.0.0.0/0\n    targetTags:\n      - http-server\n    description: \"Allow port 8080 access to http-server\"\n\n{# [END all] #}\n",
					Name:            "simple_template.jinja",
					ForceSendFields: nil,
					NullFields:      nil,
				},
				{
					Content:         "# Copyright 2016 Google Inc. All rights reserved.\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n\n# [START all]\n\ninfo:\n  title: Bookshelf GCE Deploy\n  author: Google Inc.\n  description: Creates a GCE Deployment\n\nrequired:\n- zone\n- machine-type\n- min-instances\n- max-instances\n- scopes\n\nproperties:\n  zone:\n    description: Zone to create the resources in.\n    type: string\n  machine-type:\n    description: Type of machine to use\n    type: string\n  machine-image:\n    description: The OS image to use on the machines\n    type: string\n  min-instances:\n    description: The minimum number of VMs the autoscaler will create\n    type: integer\n  max-instances:\n    description: The maximum number of VMs the autoscaler will create\n    type: integer\n  target-utilization:\n    description: The target CPU usage for the autoscaler to base its scaling on\n    type: number\n  scopes:\n    description: A list of scopes to create the VM with\n    type: array\n    minItems: 1\n    items:\n      type: string\n\n# [END all]\n",
					Name:            "",
					ForceSendFields: nil,
					NullFields:      nil,
				},
			},
		},
	}
	dm := new(MockDeploymentService)
	d := DeploymentmanagerService{
		Deployments: dm,
	}
	dir, _ := os.Getwd()
	err := fc.LoadConfigFileIfExists(dir, "teststack")
	if err != nil {
		t.Fatal(err)
	}
	deploymentName := "teststack"
	deployments := constructDeployment(deploymentName)
	if !reflect.DeepEqual(expectedDeployments, deployments) {
		t.Fatal("the expected deployment did not match the got deployments")
	}
	err = insertDeployments(d, deployments, deploymentName)
	if err == nil {
		t.Fatal("was expecting error. got nothing.")
	}
	if err.Error() != "return value was nil" {
		t.Fatal("wrong error message. got: ", err.Error())
	}
}
