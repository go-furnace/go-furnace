package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Skarlso/go-furnace/furnace-aws/plugins/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion: 1,
	MagicCookieKey:  "FURNACE_PLUGINS",
	// Never ever change this.
	MagicCookieValue: "5f7fcb61-90a3-4a90-92d1-06c8eabd20e4",
}

// PreCreateGRPCPlugin is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type PreCreateGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl PreCreate
}

// GRPCPreCreateClient is an implementation of PreCreate that talks over RPC.
type GRPCPreCreateClient struct{ client proto.PreCreateClient }

// Execute is the GRPC implementation of the Execute function for the
// PreCreate plugin definition. This will talk over GRPC.
func (m *GRPCPreCreateClient) Execute(key string) bool {
	p, err := m.client.Execute(context.Background(), &proto.Stack{
		Name: key,
	})
	if err != nil {
		return false
	}
	return p.Failed
}

// GRPCPreCreateServer is the gRPC server that GRPCPreCreateClient talks to.
type GRPCPreCreateServer struct {
	// This is the real implementation
	Impl PreCreate
}

// GRPCServer is the grpc server implementation which calls the
// protoc generated code to register it.
func (p *PreCreateGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPreCreateServer(s, &GRPCPreCreateServer{Impl: p.Impl})
	return nil
}

// GRPCClient is the grpc client that will talk to the GRPC Server
// and calls into the generated protoc code.
func (p *PreCreateGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPreCreateClient{client: proto.NewPreCreateClient(c)}, nil
}

// Execute is the execute functin of the GRPCServer which will rely the information to the
// underlying implementation of this interface.
func (m *GRPCPreCreateServer) Execute(ctx context.Context, req *proto.Stack) (*proto.Proceed, error) {
	res := m.Impl.Execute(req.Name)
	return &proto.Proceed{Failed: res}, nil
}

// PostCreate interface is the definition of the PostCreate api that can be
// implemented and used via plugins. This interface gives access to the
// stack name.
type PostCreate interface {
	Execute(key string)
}

// PreCreate is the interface for anything before the build happens. The
// PreCreate plugin has the change to abort the build if returns false.
type PreCreate interface {
	Execute(key string) bool
}

// RunPreCreatePlugins will execute all the PreCreate plugins. This function
// uses plugin discovery via the glob:
// PreCreate plugins: `*-furnace-precreate`
func RunPreCreatePlugins(stackname string) {
	// TODO: Fill the plugin map with the names of the plugins..?
	ps, _ := discoverPlugins("*-furnace-precreate")
	pluginMap := make(map[string]plugin.Plugin, 0)
	for _, v := range ps {
		pluginName := filepath.Base(v)
		pluginMap[pluginName] = &PreCreateGRPCPlugin{}
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

		p := raw.(PreCreate)
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

}

func discoverPlugins(postfix string) (p []string, err error) {
	plugs, err := plugin.Discover(postfix, "./plugins")
	if err != nil {
		return nil, err
	}
	fmt.Println("Plugins found: ", plugs)
	return plugs, nil
}
