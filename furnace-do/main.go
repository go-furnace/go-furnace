package main

import (
	"fmt"

	cmd "github.com/Yitsushi/go-commander"
	"github.com/go-furnace/go-furnace/furnace-do/commands"
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
 ________      ______
|"      "\    /    " \
(.  ___  :)  // ____  \
|: \   ) || /  /    ) :)
(| (___\ ||(: (____/ //
|:       :) \        /
(________/   \"_____/
	`)
	registry := cmd.NewCommandRegistry()
	registry.Register(commands.NewCreate)
	registry.Execute()
}
