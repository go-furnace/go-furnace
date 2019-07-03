package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-furnace/go-furnace/config"
	"github.com/go-furnace/go-furnace/handle"
	"gopkg.in/yaml.v2"
)

// Configuration object with all the properties that AWS needs.
type Configuration struct {
	Main struct {
		Stackname string `yaml:"stackname"`
		Spinner   int    `yaml:"spinner"`
		Plugins   struct {
			PluginPath string `yaml:"plugin_path"`
		} `yaml:"plugins"`
		UseDefaults bool `yaml:"use_defaults,omitempty"`
	} `yaml:"main"`
	Aws struct {
		CodeDeployRole string `yaml:"code_deploy_role"`
		Region         string `yaml:"region"`
		TemplateName   string `yaml:"template_name"`
		AppName        string `yaml:"app_name"`
		CodeDeploy     struct {
			S3Bucket    string `yaml:"code_deploy_s3_bucket,omitempty"`
			S3Key       string `yaml:"code_deploy_s3_key,omitempty"`
			GitAccount  string `yaml:"git_account,omitempty"`
			GitRevision string `yaml:"git_revision,omitempty"`
		} `yaml:"code_deploy"`
	} `yaml:"aws"`
}

// Config is the loaded configuration entity.
var Config Configuration

// RunPlugin is a plugin to execute
type RunPlugin struct {
	Run  interface{}
	Name string
}

var configPath string
var templatePath string

var defaultConfig = "aws_furnace_config.yaml"

func init() {
	configPath = config.Path()
	defaultConfigPath := filepath.Join(configPath, defaultConfig)
	Config.LoadConfiguration(defaultConfigPath)
	templatePath = filepath.Join(configPath, Config.Aws.TemplateName)
}

// LoadConfiguration loads a yaml file which sets fields for Configuration struct
func (c *Configuration) LoadConfiguration(configFile string) {
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			log.Println("main configuration file does not exist. Moving on assuming a new will be defined.")
			return
		}
	}
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
	for separatorIndex > 0 {
		if _, err := os.Stat(filepath.Join(dir, "."+file+".furnace")); err == nil {
			configLocation, _ := ioutil.ReadFile(filepath.Join(dir, "."+file+".furnace"))
			configPath = dir
			Config.LoadConfiguration(filepath.Join(configPath, string(configLocation)))
			templateBase := path.Dir(string(configLocation))
			templatePath = filepath.Join(configPath, templateBase, Config.Aws.TemplateName)
			return nil
		}
		separatorIndex = strings.LastIndex(dir, string(os.PathSeparator))
		dir = dir[0:separatorIndex]
	}

	return errors.New("failed to find configuration file for stack " + file)
}

// LoadCFStackConfig Load the CF stack configuration file into a []byte.
func LoadCFStackConfig() []byte {
	dat, err := ioutil.ReadFile(templatePath)
	handle.Error(err)
	return dat
}
