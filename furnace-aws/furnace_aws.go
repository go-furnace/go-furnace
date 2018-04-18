package main

import (
	"github.com/Skarlso/go-furnace/furnace-aws/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewStatus)
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Register(commands.NewPush)
	registry.Register(commands.NewDeleteApp)
	registry.Register(commands.NewUpdate)
	registry.Execute()
}
