package config

import (
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-furnace/errorhandler"
)

var configPath string

// Configuration is a Configuration object.
type Configuration struct {
	LogLevel   string
	UploadPath string
}

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	errorhandler.CheckError(err)
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	configPath = Path()
}

// LoadCFStackConfig Load the CF stack configuration file into a []byte.
func LoadCFStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "cloud_formation.json"))
	errorhandler.CheckError(err)
	return dat
}
