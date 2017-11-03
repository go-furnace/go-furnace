package awsconfig

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"

	"strings"

	config "github.com/Skarlso/go-furnace/config/common"
)

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

// CODEDEPLOYROLE is the default name of the codedeploy role.
const CODEDEPLOYROLE = "CodeDeployServiceRole"

// REGION to operate in.
var REGION string

// Plugin is a plugin to execute
type Plugin struct {
	Run  interface{}
	Name string
}

// PluginRegistry is a registry of plugins for certain events
var PluginRegistry map[string][]Plugin

var configPath string

func init() {
	configPath = config.Path()
	REGION = os.Getenv("AWS_FURNACE_REGION")
	if len(REGION) < 1 {
		log.Fatal("Please define a region to operate in with AWS_FURNACE_REGION exp: eu-central-1.")
	}
	PluginRegistry = fillRegistry()
}

func fillRegistry() map[string][]Plugin {
	enable := os.Getenv("FURNACE_ENABLE_PLUGIN_SYSTEM")
	ret := make(map[string][]Plugin)
	if len(enable) < 1 {
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
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "cloud_formation.json"))
	config.CheckError(err)
	return dat
}
