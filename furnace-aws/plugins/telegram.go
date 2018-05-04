package main

import (
	"log"
	"os"
)

func RunPlugin(name string) {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	log.Println("using token: ", telegramToken)
	log.Println("running telegram for stack: ", name)
}
