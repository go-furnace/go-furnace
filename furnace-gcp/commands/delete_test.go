package commands

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
)

func TestDelete(t *testing.T) {
	dm := new(MockDeploymentService)
	d := DeploymentmanagerService{
		Deployments: dm,
	}
	dir, _ := os.Getwd()
	err := fc.LoadConfigFileIfExists(dir, "teststack")
	if err != nil {
		t.Fatal(err)
	}
	err = delete(d)
	if err == nil {
		t.Fatal("was expecting error. got nothing.")
	}
	if err.Error() != "return of delete was nil" {
		t.Fatal("wrong error message. got: ", err.Error())
	}
}

func TestDeleteFails(t *testing.T) {
	dm := new(MockDeploymentService)
	dm.delete = nil
	d := DeploymentmanagerService{
		Deployments: dm,
	}
	dir, _ := os.Getwd()
	err := fc.LoadConfigFileIfExists(dir, "teststack")
	if err != nil {
		t.Fatal(err)
	}
	err = delete(d)
	if err == nil {
		t.Fatal("was expecting error. got nothing")
	}
	want := "return of delete was nil"
	if err.Error() != "return of delete was nil" {
		t.Fatalf("did not get expected error of '%s'. got: %s", want, err.Error())
	}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

var getRequest = `{
  "description": "Gets information about a specific deployment.",
  "httpMethod": "GET",
  "id": "1",
  "parameterOrder": [
    "project",
    "deployment"
  ],
  "operation": {
	  "id": "1",
	  "status": "DONE"
  },
  "parameters": {
    "deployment": {
      "description": "The name of the deployment for this request.",
      "location": "path",
      "pattern": "[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?",
      "required": true,
	  "type": "string"
    },
    "project": {
      "description": "The project ID for this request.",
      "location": "path",
      "pattern": "(?:(?:[-a-z0-9]{1,63}\\.)*(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?):)?(?:[0-9]{1,19}|(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?))",
      "required": true,
      "type": "string"
	}
  },
  "path": "{project}/global/deployments/{deployment}",
  "response": {
    "$ref": "Deployment"
  },
  "scopes": [
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/cloud-platform.read-only",
    "https://www.googleapis.com/auth/ndev.cloudman",
    "https://www.googleapis.com/auth/ndev.cloudman.readonly"
  ]
}`

var deleteRequest = `
{
	"description": "Deletes a deployment and all of the resources in the deployment.",
	"httpMethod": "DELETE",
	"id": "1",
	"parameterOrder": [
	  "project",
	  "deployment"
	],
	"parameters": {
	  "deletePolicy": {
		"default": "DELETE",
		"description": "Sets the policy to use for deleting resources.",
		"enum": [
		  "ABANDON",
		  "DELETE"
		],
		"enumDescriptions": [
		  "",
		  ""
		],
		"location": "query",
		"type": "string"
	  },
	  "deployment": {
		"description": "The name of the deployment for this request.",
		"location": "path",
		"required": true,
		"type": "string"
	  },
	  "project": {
		"description": "The project ID for this request.",
		"location": "path",
		"pattern": "(?:(?:[-a-z0-9]{1,63}\\.)*(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?):)?(?:[0-9]{1,19}|(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?))",
		"required": true,
		"type": "string"
	  }
	},
	"path": "{project}/global/deployments/{deployment}",
	"response": {
	  "$ref": "Operation"
	},
	"scopes": [
	  "https://www.googleapis.com/auth/cloud-platform",
	  "https://www.googleapis.com/auth/ndev.cloudman"
	]
}
`

func TestDeleteExecute(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		// equals(t, req.URL.String(), "http://example.com/some/path")
		if req.Method == "DELETE" {
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(deleteRequest)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		} else {
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(getRequest)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		}
	})
	d := Delete{
		client: client,
	}
	opts := &commander.CommandHelper{}
	opts.Args = make([]string, 0)
	opts.Args = append(opts.Args, "teststack")
	d.Execute(opts)
}
