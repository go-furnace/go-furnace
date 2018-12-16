package main

import (
	"fmt"

	cmd "github.com/Yitsushi/go-commander"
	"github.com/go-furnace/go-furnace/furnace-aws/commands"
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
     __       __   __  ___   ________
    /""\     |"  |/  \|  "| /"       )
   /    \    |'  /    \:  |(:   \___/
  /' /\  \   |: /'        | \___  \
 //  __'  \   \//  /\'    |  __/  \\
/   /  \\  \  /   /  \\   | /" \   :)
(___/    \___)|___/    \___|(_______/
	`)
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
