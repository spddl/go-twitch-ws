package twitch

import (
	"fmt"
	"strings"
)

func (c *Client) Login() {
	c.emitQueue.Authenticate <- "CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"
	if !strings.HasPrefix(c.User, "justinfan") {
		c.emitQueue.Authenticate <- fmt.Sprintf("PASS oauth:%s", c.Oauth)
	}
	c.emitQueue.Authenticate <- fmt.Sprintf("NICK %s", c.User)
}

// Join accept channels only in lowercase
func (c *Client) Join(channels []string) {
	for _, channel := range channels {
		c.emitQueue.Join <- fmt.Sprintf(":%s! JOIN #%s", c.User, channel)
	}
	// https://github.com/gempir/go-twitch-irc/issues/102#issuecomment-510882229
}

// Part accept channels only in lowercase
func (c *Client) Part(channels []string) {
	for _, channel := range channels {
		c.emitQueue.Join <- fmt.Sprintf(":%s! PART #%s", c.User, channel)
	}
}

// Say channel without #
func (c *Client) Say(channel, msg string, modPrivileged bool) {
	if modPrivileged {
		c.emitQueue.ModOp <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #%s :%s", channel, msg)
	} else {
		c.emitQueue.RateLimit <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #%s :%s", channel, msg)
	}

}

// Whisper TODO: currently without rate limit
func (c *Client) Whisper(nick, msg string) {
	c.emitQueue.Whisper <- fmt.Sprintf(":tmi.twitch.tv PRIVMSG #jtv :/w %s %s", nick, msg) // TODO Rate Limit
}
