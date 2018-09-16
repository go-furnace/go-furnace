package main

import (
	"log"

	"github.com/Skarlso/furnace-gosdk"
	fplugins "github.com/Skarlso/go-furnace/furnace-aws/plugins"
	"github.com/hashicorp/go-plugin"
)

// TelegramPreCreate is an actual implementation of the furnace PreCreate plugin
// interface.
type TelegramPreCreate struct{}

// Execute is the entry point to this plugin.
func (TelegramPreCreate) Execute(stackname string) bool {
	log.Println("got stackname: ", stackname)
	return false
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: fplugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			"telegram-furnace-precreate": &gosdk.PreCreateGRPCPlugin{Impl: &TelegramPreCreate{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
