package commands

import (
	"os"
	"testing"

	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
)

func TestDelete(t *testing.T) {
	dm := new(MockDeploymentService)
	d := NewDeploymentService(nil, dm)
	dir, _ := os.Getwd()
	fc.LoadConfigFileIfExists(dir, "teststack")
	delete(d)
}
