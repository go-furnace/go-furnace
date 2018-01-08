package googleconfig

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	config "github.com/Skarlso/go-furnace/config"
)

var configPath string

// GOOGLEPROJECTNAME The name of the google project to do the deployment in.
var GOOGLEPROJECTNAME string

func init() {
	GOOGLEPROJECTNAME = os.Getenv("GOOGLE_PROJECT_NAME")
	configPath = config.Path()
}

// LoadGoogleStackConfig Loads the google stack configuration file.
func LoadGoogleStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "google_template.yaml"))
	config.CheckError(err)
	return dat
}

// LoadImportFileContent Load import file contents.
func LoadImportFileContent(name string) []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, name))
	config.CheckError(err)
	return dat
}

// LoadSchemaForPath returns the content of possible schema files.
func LoadSchemaForPath(name string) (bool, []byte) {
	schema := filepath.Join(configPath, name+".schema")
	log.Println("Looking for schema file for: ", name)
	log.Println("Schema to look for is: ", schema)
	if _, err := os.Stat(schema); os.IsNotExist(err) {
		return false, []byte{}
	}
	dat, err := ioutil.ReadFile(schema)
	config.CheckError(err)
	return true, dat
}
