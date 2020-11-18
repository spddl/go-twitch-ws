package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"

	"github.com/spddl/go-twitch-ws"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	bot, err := twitch.NewClient(&twitch.Client{
		Server:      "wss://irc-ws.chat.twitch.tv", // SSL, without SSL: ws://irc-ws.chat.twitch.tv
		User:        "",
		Oauth:       "", // without "oauth:" https://twitchapps.com/tmi/
		Debug:       true,
		BotVerified: false, // verified bots: Have higher chat limits than regular users.
		Channel:     []string{"gronkhtv", "tfue", "dreamHackcs"},
	})
	if err != nil {
		panic(err)
	}

	bot.OnPrivateMessage = func(msg twitch.IRCMessage) {
		channel := msg.Params[0][1:] // to remove # from Channel Parameter
		msgline := msg.Params[1]

		if bytes.Equal(channel, []byte("#spddl")) {
			if bytes.Contains(msgline, []byte("hi")) {
				bot.Say("spddl", "Hi!", false) // only with creds
			}
		}

		log.Printf("%s - %s: %s", msg.Params[0][1:], msg.Tags["display-name"], msg.Params[1])
	}

	bot.Run()

	for { // ctrl - c
		<-interrupt
		bot.Close()
		os.Exit(0)
	}
}
