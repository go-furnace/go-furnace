package main

import (
	"fmt"

	cmd "github.com/Yitsushi/go-commander"
	"github.com/go-furnace/go-furnace/furnace-gcp/commands"
)

func main() {
	fmt.Println(`
  _______  ____  ____   _______   _____  ___        __       ______    _______
 /"     "|("  _||_ " | /"      \ (\"   \|"  \      /""\     /" _  "\  /"     "|
(: ______)|   (  ) : ||:        ||.\\   \    |    /    \   (: ( \___)(: ______)
 \/    |  (:  |  | . )|_____/   )|: \.   \\  |   /' /\  \   \/ \      \/    |
 // ___)   \\ \__/ //  //      / |.  \    \. |  //  __'  \  //  \ _   // ___)_
(:  (      /\\ __ //\ |:  __   \ |    \    \ | /   /  \\  \(:   _) \ (:      "|
 \__/     (__________)|__|  \___) \___|\____\)(___/    \___)\_______) \_______)
	`)
	fmt.Println(`
  _______    ______     _______
 /" _   "|  /" _  "\   |   __ "\
(: ( \___) (: ( \___)  (. |__) :)
 \/ \       \/ \       |:  ____/
 //  \ ___  //  \ _    (|  /
(:   _(  _|(:   _) \  /|__/ \
 \_______)  \_______)(_______)
 	`)
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Register(commands.NewDelete)
	registry.Register(commands.NewStatus)
	registry.Register(commands.NewUpdate)
	registry.Execute()
}
