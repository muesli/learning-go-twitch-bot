package main

import (
	"fmt"
	"log"
	"os"

	bot "github.com/curi0s/learning-go-twitch-bot"
)

func bla(t *bot.Twitch, ev bot.EventMessageReceived) {
	fmt.Println(ev.User, ev.Message)
}

func handleEvents(t *bot.Twitch, ch chan interface{}) error {
	for event := range ch {
		switch ev := event.(type) {
		case bot.EventConnected:
			log.Println("Connected!")

		case bot.EventPinged:
			log.Println("PING!")

		case bot.EventMessageReceived:
			go bla(t, ev)

		case bot.ConnectionError:
			return ev.Err
		}
	}

	return nil
}

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("Empty TOKEN")
	}

	t := bot.NewTwitch(bot.Options{
		Username: "cereal_bot",
		Token:    token,
		Channel:  "#codecereal",
	})

	ch := t.Connect()
	err := handleEvents(t, ch)
	if err != nil {
		log.Fatal(err)
	}
}
