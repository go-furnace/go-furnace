package plugins

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	awsconfig "github.com/go-furnace/go-furnace/furnace-aws/config"
	"github.com/go-furnace/go-furnace/handle"
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
	pluginMap := make(map[string]plugin.Plugin)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PreCreateGRPCPlugin{}
	}

	for _, v := range ps {
		raw, client := getRawAndClientForPlugin(pluginMap, v)

		p := raw.(sdk.PreCreate)
		ret := p.Execute(stackname)
		client.Kill()
		if !ret {
			log.Printf("A plugin with name '%s' prevented create to run.\n", v)
			err := errors.New("plugin prevented create to run")
			handle.Fatal(err.Error(), err)
		}
	}
}

// RunPostCreatePlugins will execute all the PreCreate plugins. This function
// uses plugin discovery via the glob:
// PostCreate plugins: `*-furnace-postcreate`
func RunPostCreatePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-postcreate*")
	pluginMap := make(map[string]plugin.Plugin)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PostCreateGRPCPlugin{}
	}

	for _, v := range ps {
		raw, client := getRawAndClientForPlugin(pluginMap, v)

		p := raw.(sdk.PostCreate)
		p.Execute(stackname)
		client.Kill()
	}
}

// RunPreDeletePlugins will execute all the PreDelete plugins. This function
// uses plugin discovery via the glob:
// PreDelete plugins: `*-furnace-predelete*`
func RunPreDeletePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-predelete*")
	pluginMap := make(map[string]plugin.Plugin)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PreDeleteGRPCPlugin{}
	}

	for _, v := range ps {
		raw, client := getRawAndClientForPlugin(pluginMap, v)

		p := raw.(sdk.PreDelete)
		ret := p.Execute(stackname)
		client.Kill()
		if !ret {
			log.Printf("A plugin with name '%s' prevented delete to run.\n", v)
			err := errors.New("plugin prevented delete to run")
			handle.Fatal(err.Error(), err)
		}
	}
}

// RunPostDeletePlugins will execute all the PostDelete plugins. This function
// uses plugin discovery via the glob:
// PostDelete plugins: `*-furnace-postdelete`
func RunPostDeletePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-postdelete*")
	pluginMap := make(map[string]plugin.Plugin)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &sdk.PostDeleteGRPCPlugin{}
	}

	for _, v := range ps {
		raw, client := getRawAndClientForPlugin(pluginMap, v)

		p := raw.(sdk.PostDelete)
		p.Execute(stackname)
		client.Kill()
	}
}

func getRawAndClientForPlugin(pluginMap map[string]plugin.Plugin, v string) (interface{}, *plugin.Client) {
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
	return raw, client
}

func discoverPlugins(postfix string) (p []string, err error) {
	plugs, err := plugin.Discover(postfix, awsconfig.Config.Main.Plugins.PluginPath)
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
