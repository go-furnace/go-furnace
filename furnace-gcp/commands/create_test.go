package commands

import (
	"os"
	"testing"

	fc "github.com/Skarlso/go-furnace/furnace-gcp/config"
	dm "google.golang.org/api/deploymentmanager/v2"
)

type MockDeploymentService struct {
}

func (m *MockDeploymentService) Insert(project string, deployment *dm.Deployment) *dm.DeploymentsInsertCall {
	return nil
}

func TestExecute(t *testing.T) {
	dm := new(MockDeploymentService)
	d := NewDeploymentService(nil, dm)
	dir, _ := os.Getwd()
	fc.LoadConfigFileIfExists(dir, "teststack")
	deploymentName := "teststack"
	deployments := constructDeploymen(deploymentName)
	err := insertDeployments(d, deployments, deploymentName)
	if err == nil {
		t.Fatal("was expecting error. got nothing.")
	}
	if err.Error() != "return value was nil" {
		t.Fatal("wrong error message. got: ", err.Error())
	}
}
