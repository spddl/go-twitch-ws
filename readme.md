# go-twitch-ws

I wanted to build a Twitch Bouncer in Go but I was missing the Websocket interface
I only found this: https://github.com/gempir/go-twitch-irc
I also build my own light variant for the twitch-websockets interface

### Help/Ideas/Improvements from

[Twitch Docs](https://dev.twitch.tv/docs/irc/guide#connecting-to-twitch-irc)
[Getting started with websocket client in go](https://medium.com/@sachinshinde7676/getting-started-with-websocket-client-in-go-73baaf8b5caf)
[go-twitch-irc](https://github.com/gempir/go-twitch-irc)
[recws](https://github.com/recws-org/recws)
[irc-message parser](https://github.com/sigkell/irc-message)
[tmi.js](https://github.com/tmijs/tmi.js)
[twitch-js](https://github.com/twitch-js/twitch-js)

## Documentation

https://pkg.go.dev/github.com/spddl/go-twitch-ws

## Getting Started
```go
interrupt := make(chan os.Signal, 1)
signal.Notify(interrupt, os.Interrupt)

bot, err := twitch.NewClient(twitch.Client{
  Server:      "wss://irc-ws.chat.twitch.tv", // SSL, without SSL: ws://irc-ws.chat.twitch.tv
  User:        "",
  Oauth:       "", // without "oauth:" https://twitchapps.com/tmi/
  Debug:       true,
  BotVerified: false, // verified bots: Have higher chat limits than regular users.
  Channel: []string{"gronkhtv", "tfue", "dreamhackcs"}, // only in Lowercase
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
  log.Println(fmt.Sprintf("%s - %s: %s", msg.Params[0][1:], msg.Tags["display-name"], msg.Params[1]))
}

bot.Run()

for { // ctrl - c
  <-interrupt
  bot.Close()
  os.Exit(0)
}
```
