package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Skarlso/go-furnace/furnace-aws/plugins/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"slack-prebuild": &PreBuildGRPCPlugin{},
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "FURNACE_PLUGINS",
	MagicCookieValue: "lkjasdfkjhasdfljksaalkajfdioh",
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type PreBuildGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl PreBuild
}

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCPreBuildClient struct{ client proto.PreBuildClient }

func (m *GRPCPreBuildClient) Execute(key string) bool {
	p, err := m.client.Execute(context.Background(), &proto.Stack{
		Name: key,
	})
	if err != nil {
		return false
	}
	return p.Failed
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCPreBuildServer struct {
	// This is the real implementation
	Impl PreBuild
}

func (p *PreBuildGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPreBuildServer(s, &GRPCPreBuildServer{Impl: p.Impl})
	return nil
}

func (p *PreBuildGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPreBuildClient{client: proto.NewPreBuildClient(c)}, nil
}

func (m *GRPCPreBuildServer) Execute(ctx context.Context, req *proto.Stack) (*proto.Proceed, error) {
	res := m.Impl.Execute(req.Name)
	return &proto.Proceed{Failed: res}, nil
}

type PostBuild interface {
	Execute(key string)
}

type PreBuild interface {
	Execute(key string) bool
}

func RunPreBuildPlugins(stackname string) {
	// TODO: Fill the plugin map with the names of the plugins..?
	discoverPreBuildPlugins()

	// TODO: Exec command should be the names from discoveredPreBuildPlugins...?
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command("./plugins/slack-plugin"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC},
	})
	defer client.Kill()
	grpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := grpcClient.Dispense("slack-prebuild")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	p := raw.(PreBuild)
	ret := p.Execute(stackname)
	if !ret {
		fmt.Println("Plugin said NO!")
		os.Exit(1)
	}
}

func RunPostBuildPlugins() {

}

func discoverPreBuildPlugins() (p []string, err error) {
	plugs, err := plugin.Discover("*-furnace-prebuild", "./plugins")
	if err != nil {
		return nil, err
	}
	fmt.Println("Plugins found: ", plugs)
	return plugs, nil
}

func discoverPostBuildPlugins() {

}
