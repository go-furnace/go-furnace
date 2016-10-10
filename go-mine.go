package main

import (
	"github.com/Skarlso/go_aws_mine/commands"
	cmd "github.com/Yitsushi/go-commander"
)

func main() {
	registry := cmd.NewCommandRegistry()
	registry.Register("create-ec2", &commands.CreateEC2{})
	registry.Execute()
}
