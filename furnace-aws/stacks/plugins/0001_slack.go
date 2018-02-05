package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

func RunPlugin() {
	stackname := "FurnaceStack"
	api := slack.New("YOUR_TOKEN_HERE")
	params := slack.PostMessageParameters{}
	channelID, timestamp, err := api.PostMessage("#general", fmt.Sprintf("Stack with name '%s' is Done.", stackname), params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
