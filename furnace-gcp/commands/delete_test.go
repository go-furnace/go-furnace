package commands

import (
	"os"
	"testing"

	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
)

func TestDelete(t *testing.T) {
	dm := new(MockDeploymentService)
	d := DeploymentmanagerService{
		Deployments: dm,
	}
	dir, _ := os.Getwd()
	fc.LoadConfigFileIfExists(dir, "teststack")
	err := delete(d)
	if err == nil {
		t.Fatal("was expecting error. got nothing.")
	}
	if err.Error() != "return of delete was nil" {
		t.Fatal("wrong error message. got: ", err.Error())
	}
}
