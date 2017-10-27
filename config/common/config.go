package commonconfig

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const (
	// PRECREATE Event name for plugins
	PRECREATE = "pre_create"
	// POSTCREATE Event name for plugins
	POSTCREATE = "post_create"
	// PREDELETE Event name for plugins
	PREDELETE = "pre_delete"
	// POSTDELETE Event name for plugins
	POSTDELETE = "post_delete"
)

// WAITFREQUENCY global wait frequency default.
var WAITFREQUENCY = 1

// STACKNAME is the default name for a stack.
var STACKNAME = "FurnaceStack"

// SPINNER is the index of which spinner to use. Defaults to 7.
var SPINNER int

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	stackname := os.Getenv("FURNACE_STACKNAME")
	if len(stackname) > 0 {
		STACKNAME = stackname
	}
}
