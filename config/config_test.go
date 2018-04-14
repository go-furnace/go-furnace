package config

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/Skarlso/go-furnace/handle"
)

// TestConfigPath Test configuration path loader.
func TestConfigPathQuick(t *testing.T) {
	usr, _ := user.Current()
	actualConfigPath := Path()
	expectedConfigPath := filepath.Join(usr.HomeDir, ".config", "go-furnace")
	fmt.Println("Config path is: ", actualConfigPath)
	if actualConfigPath != expectedConfigPath {
		t.Fatalf("Expected: %s != Actual %s.", expectedConfigPath, actualConfigPath)
	}
}

func TestCheckError(t *testing.T) {
	failed := false
	handle.LogFatalf = func(format string, v ...interface{}) {
		failed = true
	}
	err := errors.New("test error")
	handle.Error(err)
	if !failed {
		t.Fatal("Should have failed.")
	}
}

func TestHandleFatal(t *testing.T) {
	failed := false
	handle.LogFatalf = func(format string, v ...interface{}) {
		failed = true
	}
	err := errors.New("test error")
	handle.Fatal("format", err)
	if !failed {
		t.Fatal("Should have failed.")
	}
}
