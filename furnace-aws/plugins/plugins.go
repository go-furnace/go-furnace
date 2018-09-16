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
		pluginMap[pluginName] = &gosdk.PreCreateGRPCPlugin{}
	}

	for _, v := range ps {
		var cmd *exec.Cmd
		ext := filepath.Ext(v)
		switch ext {
		case ".py":
			python, err := exec.LookPath("python3")
			if err != nil {
				log.Println("Could not locate binary for python3 on PATH.")
				os.Exit(1)
			}
			cmd = exec.Command(python, v)
		case ".rb":
			ruby, err := exec.LookPath("ruby")
			if err != nil {
				log.Println("Could not locate binary for ruby on PATH.")
				os.Exit(1)
			}
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

		p := raw.(gosdk.PreCreate)
		ret := p.Execute(stackname)
		if !ret {
			fmt.Println("Plugin said NO!")
			os.Exit(1)
		}
	}
}

// RunPostCreatePlugins will execute all the PreCreate plugins. This function
// uses plugin discovery via the glob:
// PostCreate plugins: `*-furnace-postcreate`
func RunPostCreatePlugins(stackname string) {
	ps, _ := discoverPlugins("*-furnace-postcreate")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &gosdk.PostCreateGRPCPlugin{}
	}

	for _, v := range ps {
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins:         pluginMap,
			Cmd:             exec.Command(v),
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

		p := raw.(gosdk.PostCreate)
		p.Execute(stackname)
	}
}

func discoverPlugins(postfix string) (p []string, err error) {
	plugs, err := plugin.Discover(postfix, "./plugins")
	if err != nil {
		return nil, err
	}
	fmt.Println("Plugins found: ", plugs)
	return plugs, nil
}
