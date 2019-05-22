package config

import (
	"os/user"
	"path/filepath"

	"github.com/go-furnace/go-furnace/handle"
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

// SPINNER is the index of which spinner to use. Defaults to 7.
var SPINNER = 7

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	handle.Error(err)
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}
