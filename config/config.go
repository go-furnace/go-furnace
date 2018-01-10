package config

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

// Spinners is a collection os spinner types
var Spinners = []string{`←↖↑↗→↘↓↙`,
	`▁▃▄▅▆▇█▇▆▅▄▃`,
	`┤┘┴└├┌┬┐`,
	`◰◳◲◱`,
	`◴◷◶◵`,
	`◐◓◑◒`,
	`⣾⣽⣻⢿⡿⣟⣯⣷`,
	`|/-\`}

// WAITFREQUENCY global wait frequency default.
var WAITFREQUENCY = 1

// STACKNAME is the default name for a stack.
var STACKNAME = "FurnaceStack"

// SPINNER is the index of which spinner to use. Defaults to 7.
var SPINNER = 7

// LogFatalf is the function to log a fatal error.
var LogFatalf = log.Fatalf

// CheckError handles errors.
func CheckError(err error) {
	if err != nil {
		HandleFatal("Error occurred:", err)
	}
}

// HandleFatal handler fatal errors in Furnace.
func HandleFatal(s string, err error) {
	LogFatalf(s, err)
}

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	CheckError(err)
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	stackname := os.Getenv("FURNACE_STACKNAME")
	if len(stackname) > 0 {
		STACKNAME = stackname
	}
}
