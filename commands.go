package twitch

import (
	"fmt"
	"log"
	"strings"
)

func (client *Client) Login() {
	log.Println("Login()")

	// client.write([]byte("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"))
	// if !strings.HasPrefix(client.User, "justinfan") {
	// 	client.write([]byte(fmt.Sprintf("PASS oauth:%s", client.Oauth)))
	// }
	// client.write([]byte(fmt.Sprintf("NICK %s", client.User)))

	client.emitQueue.Authenticate <- "CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"
	if !strings.HasPrefix(client.User, "justinfan") {
		client.emitQueue.Authenticate <- fmt.Sprintf("PASS oauth:%s", client.Oauth)
	}
	client.emitQueue.Authenticate <- fmt.Sprintf("NICK %s", client.User)
}

// Join accept channels only in lowercase
func (client *Client) Join(channels []string) {
	for _, channel := range channels {
		client.emitQueue.Join <- fmt.Sprintf(":%s! JOIN #%s", client.User, channel)
	}
	// https://github.com/gempir/go-twitch-irc/issues/102#issuecomment-510882229
}

// Part accept channels only in lowercase
func (client *Client) Part(channels []string) {
	for _, channel := range channels {
		client.emitQueue.Join <- fmt.Sprintf(":%s! PART #%s", client.User, channel)
	}
}

// Say channel without #
func (client *Client) Say(channel, msg string, modPrivileged bool) {
	if modPrivileged {
		client.emitQueue.ModOp <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #%s :%s", channel, msg)
	} else {
		client.emitQueue.RateLimit <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #%s :%s", channel, msg)
	}

}

func (client *Client) Whisper(nick, msg string) {
	client.emitQueue.RateLimit <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #jtv :/w %s %s", nick, msg)
}
