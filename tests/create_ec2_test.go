package tests

import (
	"log"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/Skarlso/go-aws-mine/config"
)

func TestCreatingAnEC2InstanceTest(t *testing.T) {
	// configPath := Pat
	usr, _ := user.Current()
	actualConfigPath := config.Path()
	expectedConfigPath := filepath.Join(usr.HomeDir, ".config", "go-aws-mine")
	log.Println("Config path is: ", actualConfigPath)
	if actualConfigPath != expectedConfigPath {
		t.Fatalf("Expected: %s != Actual %s.", expectedConfigPath, actualConfigPath)
	}

}
