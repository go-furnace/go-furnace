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
