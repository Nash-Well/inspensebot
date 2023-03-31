package main

import (
	"inspense-bot/bot"
	"log"
)

func main() {
	b, err := bot.New("bot.yml")
	if err != nil {
		log.Fatal(err)
	}

	b.Start()
}
