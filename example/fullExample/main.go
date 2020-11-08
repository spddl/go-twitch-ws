package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

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
		BotVerified: false,                                       // verified bots: Have higher chat limits than regular users.
		Channel:     []string{"gronkhtv", "tfue", "dreamHackcs"}, // only in Lowercase
	})
	if err != nil {
		panic(err)
	}

	bot.OnPrivateMessage = func(msg twitch.IRCMessage) {
		channel := msg.Params[0][1:] // to remove # from Channel Parameter
		msgline := msg.Params[1]

		// 		log.Println(fmt.Sprintf(`{
		// 	"Raw": "%s",
		// 	"Tags": {
		// 		"badge-info": "%s",
		// 		"badges": "%s",
		// 		"color": "%s",
		// 		"display-name": "%s",
		// 		"emotes": "%s",
		// 		"flags": "%s",
		// 		"id": "%s",
		// 		"mod": "%s",
		// 		"room-id": "%s",
		// 		"subscriber": "%s",
		// 		"tmi-sent-ts": "%s",
		// 		"turbo": "%s",
		// 		"user-id": "%s",
		// 		"user-type": "%s"
		// 	},
		// 	"Command": "%s",
		// 	"Params": [
		// 		"%s",
		// 		"%s"
		// 	],
		// 	"Prefix": "%s"
		// }`, msg.Raw, msg.Tags["badge-info"], msg.Tags["badges"], msg.Tags["color"], msg.Tags["display-name"], msg.Tags["emotes"], msg.Tags["flags"], msg.Tags["id"], msg.Tags["mod"], msg.Tags["room-id"], msg.Tags["subscriber"], msg.Tags["tmi-sent-ts"], msg.Tags["turbo"], msg.Tags["user-id"], msg.Tags["user-type"], msg.Command, msg.Params[0], msg.Params[1], msg.Prefix))

		if bytes.Equal(channel, []byte("#spddl")) {
			if bytes.Contains(msgline, []byte("hi")) {
				bot.Say("spddl", "Hi!", false) // only with creds
			}
		}

		log.Println(fmt.Sprintf("%s - %s: %s", msg.Params[0][1:], msg.Tags["display-name"], msg.Params[1]))
	}

	bot.OnHosttargetMessage = func(msg twitch.IRCMessage) {
		log.Printf("> HOSTTARGET, %s: %s\n", msg.Params[0][1:], msg.Params[1])
	}

	bot.OnNoticeMessage = func(msg twitch.IRCMessage) {
		log.Printf("> NOTICE, %s: %s\n", msg.Params[0][1:], msg.Params[1])
	}

	bot.OnGlobalUserSateMessage = func(msg twitch.IRCMessage) {
		log.Printf("OnGlobalUserSateMessage: %s\n", msg)
	}

	bot.OnUserStateMessage = func(msg twitch.IRCMessage) {
		log.Printf("OnUserStateMessage: %s\n", msg)
	}

	bot.OnRoomStateMessage = func(msg twitch.IRCMessage) {
		var stats []string
		for key, value := range msg.Tags {
			stats = append(stats, fmt.Sprintf("%s: %s", key, value))
		}
		log.Printf("> ROOMSTATE, %s: %s\n", msg.Params[0][1:], strings.Join(stats, ", "))
	}

	bot.OnJoinMessage = func(msg twitch.IRCMessage) {
		log.Printf("join channel: %s", msg.Params[0][1:]) // to remove # from Channel Parameter
	}

	bot.OnPartMessage = func(msg twitch.IRCMessage) {
		log.Printf("leave channel: %s", msg.Params[0][1:]) // to remove # from Channel Parameter
	}

	bot.OnUserNoticeMessage = func(msg twitch.IRCMessage) { // https://github.com/tmijs/tmi.js/blob/4bb66c433b8ae28326b4cd8567357e6ea729e91a/lib/client.js#L668
		msg.Tags["system-msg"] = bytes.ReplaceAll(msg.Tags["system-msg"], []byte{92, 115}, []byte{32}) // "\\s", " "
		msg.Tags["msg-param-sub-plan-name"] = bytes.ReplaceAll(msg.Tags["msg-param-sub-plan-name"], []byte{92, 115}, []byte{32})
		msg.Tags["msg-param-sub-plan-name"] = bytes.ReplaceAll(msg.Tags["msg-param-sub-plan-name"], []byte{92, 115}, []byte{32})

		// msgID := msg.Tags["message-type"]
		// switch string(msgID) { // https://github.com/tmijs/tmi.js/blob/4bb66c433b8ae28326b4cd8567357e6ea729e91a/lib/client.js#L680
		// // Handle resub
		// case "resub":

		// // Handle sub
		// case "sub":

		// // Handle gift sub
		// case "subgift":

		// // Handle anonymous gift sub
		// // Need proof that this event occur
		// case "anonsubgift":

		// // Handle random gift subs
		// case "submysterygift":

		// // Handle anonymous random gift subs
		// // Need proof that this event occur
		// case "anonsubmysterygift":

		// // Handle user upgrading from Prime to a normal tier sub
		// case "primepaidupgrade":

		// // Handle user upgrading from a gifted sub
		// case "giftpaidupgrade":

		// // Handle user upgrading from an anonymous gifted sub
		// case "anongiftpaidupgrade":

		// // Handle raid
		// case "raid":
		// 	var username []byte
		// 	if len(msg.Tags["msg-param-displayName"]) != 0 {
		// 		username = msg.Tags["msg-param-displayName"]
		// 	} else {
		// 		username = msg.Tags["msg-param-login"]
		// 	}
		// 	viewers := msg.Tags["msg-param-viewerCount"]

		// 	_ = username
		// 	_ = viewers

		// // Handle ritual
		// case "ritual":

		// default:

		// }

		var stats []string
		for key, value := range msg.Tags {
			stats = append(stats, fmt.Sprintf("%s: %s", key, value))
		}
		log.Printf("> USERNOTICE, %s: %s\n", msg.Params[0][1:], strings.Join(stats, ", "))
	}

	bot.OnClearChatMessage = func(msg twitch.IRCMessage) {
		channel := msg.Params[0][1:] // to remove # from Channel Parameter
		targetUser := msg.Params[1]
		banDuration := msg.Tags["ban-duration"]
		roomID := msg.Tags["room-id"]
		log.Printf("CLEARCHAT: %s(RoomID: %s) => %s (Ban Duration: %ss)\n", channel, roomID, targetUser, banDuration)
	}

	bot.OnClearMsgMessage = func(msg twitch.IRCMessage) {
		channel := msg.Params[0][1:] // to remove # from Channel Parameter
		deletetdMsg := msg.Params[1]
		targetUser := msg.Tags["login"]
		targetMsgID := msg.Tags["target-msg-id"]
		log.Printf("CLEARMSG: %s, %s: %s (Target Msg ID: %s)\n", channel, targetUser, deletetdMsg, targetMsgID)
	}

	bot.OnUnknownMessage = func(msg twitch.IRCMessage) {
		log.Printf("OnUnknownMessage: %+v\n", msg)
	}

	bot.OnPongLatency = func(msg time.Duration) {
		log.Printf("OnPongLatency: %+v\n", msg)
	}

	bot.Run()

	for { // ctrl - c
		<-interrupt
		bot.Close()
		os.Exit(0)
	}
}
