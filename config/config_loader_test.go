package config

import (
	"fmt"
	"os/user"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	usr, _ := user.Current()
	actualConfigPath := Path()
	expectedConfigPath := filepath.Join(usr.HomeDir, ".config", "go-furnace")
	fmt.Println("Config path is: ", actualConfigPath)
	if actualConfigPath != expectedConfigPath {
		t.Fatalf("Expected: %s != Actual %s.", expectedConfigPath, actualConfigPath)
	}
}
