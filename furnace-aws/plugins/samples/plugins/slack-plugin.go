package main

import (
	"log"

	fplugs "github.com/Skarlso/go-furnace/furnace-aws/plugins"
	"github.com/hashicorp/go-plugin"
)

type SlackPreBuild struct{}

func (SlackPreBuild) Execute(stackname string) bool {
	log.Println("got stackname: ", stackname)
	return true
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: fplugs.Handshake,
		Plugins: map[string]plugin.Plugin{
			"slack-prebuild": &fplugs.PreBuildGRPCPlugin{Impl: &SlackPreBuild{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
