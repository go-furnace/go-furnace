package main

import (
	cmd "github.com/Yitsushi/go-commander"
	"github.com/go-furnace/go-furnace/furnace-gcp/commands"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Execute()
}
