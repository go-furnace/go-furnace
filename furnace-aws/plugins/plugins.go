package plugins

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-furnace/sdk"
	"github.com/hashicorp/go-plugin"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion: 1,
	MagicCookieKey:  "FURNACE_PLUGINS",
	// Never ever change this.
	MagicCookieValue: "5f7fcb61-90a3-4a90-92d1-06c8eabd20e4",
}

// RunPreCreatePlugins will execute all the PreCreate plugins. This function
// uses plugin discovery via the glob:
// PreCreate plugins: `*-furnace-precreate.*`
func RunPreCreatePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-precreate*")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PreCreateGRPCPlugin{}
	}

	for _, v := range ps {
		raw := getRawForPlugin(pluginMap, v)

		p := raw.(sdk.PreCreate)
		ret := p.Execute(stackname)
		if !ret {
			log.Printf("A plugin with name '%s' prevented create to run.\n", v)
			os.Exit(1)
		}
	}
}

// RunPostCreatePlugins will execute all the PreCreate plugins. This function
// uses plugin discovery via the glob:
// PostCreate plugins: `*-furnace-postcreate`
func RunPostCreatePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-postcreate*")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PostCreateGRPCPlugin{}
	}

	for _, v := range ps {
		raw := getRawForPlugin(pluginMap, v)

		p := raw.(sdk.PostCreate)
		p.Execute(stackname)
	}
}

// RunPreDeletePlugins will execute all the PreDelete plugins. This function
// uses plugin discovery via the glob:
// PreDelete plugins: `*-furnace-predelete*`
func RunPreDeletePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-predelete*")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PreDeleteGRPCPlugin{}
	}

	for _, v := range ps {
		raw := getRawForPlugin(pluginMap, v)

		p := raw.(sdk.PreDelete)
		ret := p.Execute(stackname)
		if !ret {
			log.Printf("A plugin with name '%s' prevented delete to run.\n", v)
			os.Exit(1)
		}
	}
}

// RunPostDeletePlugins will execute all the PostDelete plugins. This function
// uses plugin discovery via the glob:
// PostDelete plugins: `*-furnace-postdelete`
func RunPostDeletePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-postdelete*")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PostDeleteGRPCPlugin{}
	}

	for _, v := range ps {
		raw := getRawForPlugin(pluginMap, v)

		p := raw.(sdk.PostDelete)
		p.Execute(stackname)
	}
}

func getRawForPlugin(pluginMap map[string]plugin.Plugin, v string) interface{} {
	var cmd *exec.Cmd
	ext := filepath.Ext(v)
	switch ext {
	case ".py":
		python := getExecutionBinary("python3")
		cmd = exec.Command(python, v)
	case ".rb":
		ruby := getExecutionBinary("ruby")
		cmd = exec.Command(ruby, v)
	default:
		cmd = exec.Command(v)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         pluginMap,
		Cmd:             cmd,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC},
	})

	defer client.Kill()
	grpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	pluginName := filepath.Base(v)
	// Request the plugin
	raw, err := grpcClient.Dispense(pluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	return raw
}

func discoverPlugins(postfix string) (p []string, err error) {
	plugs, err := plugin.Discover(postfix, "./plugins")
	if err != nil {
		return nil, err
	}
	fmt.Println("Plugins found: ", plugs)
	return plugs, nil
}

func getExecutionBinary(want string) string {
	binary, err := exec.LookPath(want)
	if err != nil {
		log.Printf("Could not locate binary for %s on PATH.\n", want)
		os.Exit(1)
	}
	return binary
}
