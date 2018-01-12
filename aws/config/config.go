package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"plugin"

	"gopkg.in/yaml.v2"

	"strings"

	config "github.com/Skarlso/go-furnace/config"
)

// TODO: Create a main config which defines a furnace config location
// This was, when running ./furnace-aws create asdf -> it would look for asdf
// as a configuration file. Like asdf_furnace_config.yaml

// Configuration object with all the properties that AWS needs.
type Configuration struct {
	Main struct {
		Stackname string `yaml:"stackname"`
		Spinner   int    `yaml:"spinner"`
	} `yaml:"main"`
	Aws struct {
		CodeDeployRole     string `yaml:"code_deploy_role"`
		Region             string `yaml:"region"`
		EnablePluginSystem bool   `yaml:"enable_plugin_system"`
		TemplateName       string `yaml:"template_name"`
		CodeDeploy         struct {
			S3Bucket    string `yaml:"code_deploy_s3_bucket,omitempty"`
			S3Key       string `yaml:"code_deploy_s3_key,omitempty"`
			GitAccount  string `yaml:"git_account,omitempty"`
			GitRevision string `yaml:"git_revision,omitempty"`
		} `yaml:"code_deploy"`
	} `yaml:"aws"`
}

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

// Config is the loaded configuration entity.
var Config Configuration

// Plugin is a plugin to execute
type Plugin struct {
	Run  interface{}
	Name string
}

// PluginRegistry is a registry of plugins for certain events
var PluginRegistry map[string][]Plugin

var configPath string
var templatePath string

var defaultConfig = "aws_furnace_config.yaml"

func init() {
	configPath = config.Path()
	templatePath = filepath.Join(configPath, defaultConfig)
	PluginRegistry = fillRegistry()
}

// LoadConfiguration loads a yaml file which sets fields for Configuration struct
func (c *Configuration) LoadConfiguration(configFile string) {
	content, err := ioutil.ReadFile(configFile)
	config.CheckError(err)
	err = yaml.Unmarshal(content, c)
	config.CheckError(err)
}

// Recusively search backwards from the current directory for a furnace config file
// with the given prefix of `file`. If found, the Configuration `Config` will be
// loaded with values gathered from the file described by that config.
// If none is found, nothing happens. The default file remains loaded.
//
// returns an error if the file is not found.
func LoadConfigFileIfExists(dir string, file string) error {
	separatorIndex := strings.LastIndex(dir, "/")
	for separatorIndex != 0 {
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

func fillRegistry() map[string][]Plugin {
	ret := make(map[string][]Plugin)
	if !Config.Aws.EnablePluginSystem {
		return ret
	}
	// log.Println("Filling plugin registry.")
	files, _ := ioutil.ReadDir(filepath.Join(configPath, "plugins"))
	pluginCount := 0
	for _, f := range files {
		split := strings.Split(f.Name(), ".")
		key := split[len(split)-1]
		fullPath := filepath.Join(configPath, "plugins", f.Name())
		p, err := plugin.Open(fullPath)
		if err != nil {
			log.Printf("Plugin '%s' failed to load. Error: %s\n", fullPath, err.Error())
			continue
		}
		run, err := p.Lookup("RunPlugin")
		if err != nil {
			log.Printf("Plugin '%s' did not have 'RunPlugin' method. Error: %s\n", fullPath, err.Error())
			continue
		}
		plug := Plugin{
			Run:  run,
			Name: f.Name(),
		}
		if p, ok := ret[key]; ok {
			p = append(p, plug)
			ret[key] = p
		} else {
			plugs := make([]Plugin, 0)
			plugs = append(plugs, plug)
			ret[key] = plugs
		}
		pluginCount++
	}
	log.Printf("'%d' plugins loaded successfully.\n", pluginCount)
	return ret
}

// LoadCFStackConfig Load the CF stack configuration file into a []byte.
func LoadCFStackConfig() []byte {
	dat, err := ioutil.ReadFile(templatePath)
	config.CheckError(err)
	return dat
}
