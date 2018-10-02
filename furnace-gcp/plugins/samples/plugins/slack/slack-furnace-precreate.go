package main

import (
	"log"

	fplugins "github.com/go-furnace/go-furnace/furnace-gcp/plugins"
	"github.com/go-furnace/sdk"
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
			"slack-furnace-precreate": &sdk.PreCreateGRPCPlugin{Impl: &SlackPreCreate{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
