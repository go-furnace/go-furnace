package main

import (
	"github.com/Skarlso/go-furnace/furnace-gcp/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Execute()
}
