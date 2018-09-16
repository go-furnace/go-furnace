package main

import (
	"log"

	"github.com/Skarlso/furnace-gosdk"
	fplugins "github.com/Skarlso/go-furnace/furnace-aws/plugins"
	"github.com/hashicorp/go-plugin"
)

// SlackPreCreate is an actual implementation of the furnace PreCreate plugin
// interface.
type SlackPreCreate struct{}

// Execute is the entry point to this plugin.
func (SlackPreCreate) Execute(stackname string) bool {
	log.Println("got stackname: ", stackname)
	return true
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: fplugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			"slack-furnace-precreate": &gosdk.PreCreateGRPCPlugin{Impl: &SlackPreCreate{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
