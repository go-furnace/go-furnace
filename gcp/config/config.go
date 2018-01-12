package googleconfig

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

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
var templateBase string

func init() {
	configPath = config.Path()
	fileName := filepath.Join(configPath, defaultConfig)
	if _, err := os.Stat(fileName); err == nil {
		Config.LoadConfiguration(fileName)
	}
	templateBase = configPath
}

// LoadConfiguration loads a yaml file which sets fields for Configuration struct
func (c *Configuration) LoadConfiguration(configFile string) {
	content, err := ioutil.ReadFile(configFile)
	config.CheckError(err)
	err = yaml.Unmarshal(content, c)
	config.CheckError(err)
}

// LoadConfigFileIfExists Search backwards from the current directory
// for a furnace config file with the given prefix of `file`. If found,
// the Configuration `Config` will be loaded with values gathered from
// the file described by that config. If none is found, nothing happens.
// The default file remains loaded.
//
// returns an error if the file is not found.
func LoadConfigFileIfExists(dir string, file string) error {
	separatorIndex := strings.LastIndex(dir, "/")
	for separatorIndex != 0 {
		if _, err := os.Stat(filepath.Join(dir, "."+file+".furnace")); err == nil {
			configLocation, _ := ioutil.ReadFile(filepath.Join(dir, "."+file+".furnace"))
			configPath = dir
			Config.LoadConfiguration(filepath.Join(configPath, string(configLocation)))
			templateBase = path.Dir(string(configLocation))
			return nil
		}
		separatorIndex = strings.LastIndex(dir, string(os.PathSeparator))
		dir = dir[0:separatorIndex]
	}

	return errors.New("failed to find configuration file for stack " + file)
}

// LoadGoogleStackConfig Loads the google stack configuration file.
func LoadGoogleStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(templateBase, Config.Gcp.TemplateName))
	config.CheckError(err)
	return dat
}

// LoadImportFileContent Load import file contents.
func LoadImportFileContent(name string) []byte {
	dat, err := ioutil.ReadFile(filepath.Join(templateBase, name))
	config.CheckError(err)
	return dat
}

// LoadSchemaForPath returns the content of possible schema files.
func LoadSchemaForPath(name string) (bool, []byte) {
	schema := filepath.Join(templateBase, name+".schema")
	log.Println("Looking for schema file for: ", name)
	log.Println("Schema to look for is: ", schema)
	if _, err := os.Stat(schema); os.IsNotExist(err) {
		return false, []byte{}
	}
	dat, err := ioutil.ReadFile(schema)
	config.CheckError(err)
	return true, dat
}
