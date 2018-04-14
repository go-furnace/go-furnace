package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/handle"
	"gopkg.in/yaml.v2"
)

// Configuration object with all the properties that GCP needs.
type Configuration struct {
	Do struct {
		TemplateName string `yaml:"template_name"`
		StackName    string `yaml:"stack_name"`
		Token        string `yaml:"token"`
	} `yaml:"do"`
}

// Config is the loaded configuration entity.
var Config Configuration

var configPath string

var defaultConfig = "do_furnace_config.yaml"
var templateBase string

func init() {
	configPath = config.Path()
	fileName := filepath.Join(configPath, defaultConfig)
	if _, err := os.Stat(fileName); err == nil {
		Config.LoadConfiguration(fileName)
	} else {
		log.Printf("WARNING: config file '%s' not found.\n", fileName)
	}
	templateBase = configPath
}

// LoadConfiguration loads a yaml file which sets fields for Configuration struct
func (c *Configuration) LoadConfiguration(configFile string) {
	content, err := ioutil.ReadFile(configFile)
	handle.Error(err)
	err = yaml.Unmarshal(content, c)
	handle.Error(err)
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

// LoadDoStackConfig Loads the digital ocean stack configuration file.
func LoadDoStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(templateBase, Config.Do.TemplateName))
	handle.Error(err)
	return dat
}
