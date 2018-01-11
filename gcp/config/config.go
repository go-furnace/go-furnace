package googleconfig

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	config "github.com/Skarlso/go-furnace/config"
	yaml "gopkg.in/yaml.v1"
)

// Configuration object with all the properties that GCP needs.
type Configuration struct {
	Main struct {
		ProjectName string `yaml:"project_name"`
		Spinner     int    `yaml:"spinner"`
	} `yaml:"main"`
	Gcp struct {
		TemplateName string `yaml:"template_name"`
		StackName    string `yaml:"stack_name"`
	} `yaml:"gcp"`
}

// Config is the loaded configuration entity.
var Config Configuration

var configPath string

var defaultConfig = "gcp_furnace_config.yaml"

func init() {
	configPath = config.Path()
	fileName := filepath.Join(configPath, defaultConfig)
	if _, err := os.Stat(fileName); err == nil {
		Config.LoadConfiguration(fileName)
	}
}

// LoadConfiguration loads a yaml file which sets fields for Configuration struct
func (c *Configuration) LoadConfiguration(configFile string) {
	content, err := ioutil.ReadFile(configFile)
	config.CheckError(err)
	err = yaml.Unmarshal(content, c)
	config.CheckError(err)
}

// LoadGoogleStackConfig Loads the google stack configuration file.
func LoadGoogleStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, Config.Gcp.TemplateName))
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
