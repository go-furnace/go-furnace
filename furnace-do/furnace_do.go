package main

import (
	cmd "github.com/Yitsushi/go-commander"
	"github.com/go-furnace/go-furnace/furnace-do/commands"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Execute()
}
