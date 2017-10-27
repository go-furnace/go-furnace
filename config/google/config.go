package googleconfig

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"strings"
)

var configPath string

// GOOGLEPROJECTNAME The name of the google project to do the deployment in.
var GOOGLEPROJECTNAME string

func init() {
	GOOGLEPROJECTNAME = os.Getenv("GOOGLE_PROJECT_NAME")
}

// LoadGoogleStackConfig Loads the google stack configuration file.
func LoadGoogleStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "google_template.yaml"))
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return dat
}

// LoadImportFileContent Load import file contents.
func LoadImportFileContent(name string) []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, name))
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return dat
}

// LoadSchemaForPath returns the content of possible schema files.
func LoadSchemaForPath(name string) (bool, []byte) {
	base := name[0:strings.LastIndex(name, ".")]
	schema := filepath.Join(configPath, base+".schema")
	log.Println("Looking for schema file for: ", name)
	log.Println("Schema to look for is: ", schema)
	if _, err := os.Stat(schema); os.IsNotExist(err) {
		return false, []byte{}
	}
	dat, err := ioutil.ReadFile(schema)
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return true, dat
}
