package main

import (
	"github.com/Skarlso/go-furnace/furnace-do/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Execute()
}
