package main

import (
	"log"
	"os"
)

func RunPlugin(name string) {
	slackToken := os.Getenv("SLACK_TOKEN")
	log.Println("using token: ", slackToken)
	log.Println("running slack for stack: ", name)
}
