// +build windows js,wasm

package twitch

import (
	"fmt"
	"strings"
)

func (c *Client) login() {
	// Membership: Adds membership state event data. By default, we do not send this data to clients without this capability. https://dev.twitch.tv/docs/irc/membership
	// Tags: Adds IRC V3 message tags to several commands, if enabled with the commands capability. https://dev.twitch.tv/docs/irc/tags
	// Commands: Enables several Twitch-specific commands. https://dev.twitch.tv/docs/irc/commands
	c.emitQueue.Authenticate <- "CAP REQ :twitch.tv/tags twitch.tv/commands"
	if !strings.HasPrefix(c.User, "justinfan") {
		c.emitQueue.Authenticate <- fmt.Sprintf("PASS oauth:%s", c.Oauth)
	}
	c.emitQueue.Authenticate <- fmt.Sprintf("NICK %s", c.User)
}

// Join accept channels only in lowercase
func (c *Client) Join(channels []string) {
	for _, channel := range channels {
		_, exist := c.channelExists(channel)
		if !exist {
			c.mu.Lock()
			c.Channel = append(c.Channel, channel)
			c.mu.Unlock()
		}
	}
	c.joinCommand(channels)
}

func (c *Client) joinCommand(channels []string) {
	for _, channel := range channels {
		c.emitQueue.Join <- fmt.Sprintf(":%s! JOIN #%s", c.User, channel)
	}
	// https://github.com/gempir/go-twitch-irc/issues/102#issuecomment-510882229
}

// Part accept channels only in lowercase
func (c *Client) Part(channels []string) {
	for _, channel := range channels {
		i, exist := c.channelExists(channel)
		if exist {
			c.mu.Lock()
			// c.Channel[i] = c.Channel[len(c.Channel)-1]
			// c.Channel[len(c.Channel)-1] = ""
			// c.Channel = c.Channel[:len(c.Channel)-1]
			c.Channel = append(c.Channel[:i], c.Channel[i+1:]...)
			c.mu.Unlock()
		}
	}
	c.partCommand(channels)
}

func (c *Client) partCommand(channels []string) {
	for _, channel := range channels {
		c.emitQueue.Join <- fmt.Sprintf(":%s! PART #%s", c.User, channel)
	}
}

func (c *Client) channelExists(channel string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for index, item := range c.Channel {
		if item == channel {
			return index, true
		}
	}
	return -1, false
}

const sayTemplate = ":tmi.twitch.tv PRIVMSG #%s :%s"

// Say channel without #
func (c *Client) Say(channel, msg string, modPrivileged bool) {
	if modPrivileged {
		c.emitQueue.ModOp <- fmt.Sprintf(sayTemplate, channel, msg)
	} else {
		c.emitQueue.RateLimit <- fmt.Sprintf(sayTemplate, channel, msg)
	}
}

const whisperTemplate = ":tmi.twitch.tv PRIVMSG #jtv :/w %s %s"

func (c *Client) Whisper(nick, msg string) {
	if c.BotVerified {
		c.emitQueue.WhisperVerifiedBots <- fmt.Sprintf(whisperTemplate, nick, msg)
	} else if c.BotKnown {
		c.emitQueue.WhisperKnownBot <- fmt.Sprintf(whisperTemplate, nick, msg)
	} else {
		c.emitQueue.Whisper <- fmt.Sprintf(whisperTemplate, nick, msg)
	}
}
