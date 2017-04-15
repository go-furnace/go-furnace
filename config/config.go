package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"plugin"
	"strconv"

	"strings"
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

var configPath string

// WAITFREQUENCY global wait frequency default.
var WAITFREQUENCY = 1

// STACKNAME is the default name for a stack.
var STACKNAME = "FurnaceStack"

// SPINNER is the index of which spinner to use. Defaults to 7.
var SPINNER int

// Plugin is a plugin to execute
type Plugin struct {
	Run  interface{}
	Name string
}

// PluginRegistry is a registry of plugins for certain events
var PluginRegistry map[string][]Plugin

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	configPath = Path()
	REGION = os.Getenv("FURNACE_REGION")
	spinner := os.Getenv("FURNACE_SPINNER")
	if len(spinner) < 1 {
		SPINNER = 7
	} else {
		SPINNER, _ = strconv.Atoi(spinner)
	}
	if len(REGION) < 1 {
		log.Fatal("Please define a region to operate in with FURNACE_REGION exp: eu-central-1.")
	}
	stackname := os.Getenv("FURNACE_STACKNAME")
	if len(stackname) > 0 {
		STACKNAME = stackname
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
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return dat
}
