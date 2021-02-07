// +build windows linux js,wasm

package twitch

import (
	"bytes"
	"log"
	"time"
)

func (c *Client) write(msg []byte) {
	select {
	case <-c.context.Done():
		c.Close()
		return
	default:
		if c.IsConnected() {
			w, err := c.conn.Writer(c.context, 1 /* websocket.MessageText */)
			if err != nil {
				c.CloseAndReconnect()
				return
			}

			_, err = w.Write(append(msg, []byte{13, 10}...))
			if err != nil {
				log.Println(err)
				return
			}

			err = w.Close()
			if err != nil {
				log.Println(err)
				return
			}
			if c.Debug {
				log.Printf(debugTemplate, msg)
			}
		}
	}
}

func (c *Client) parser(msgData []byte) {
	msg := bytes.Split(msgData, []byte{13, 10}) // "\r\n"
	for _, v := range msg {
		if len(v) == 0 {
			continue
		}

		// log.Println(string(v))

		ircMsg, err := parseIRCMessage(v)
		if err != nil {
			log.Println("parseIRCMessage:", err)
			return
		}

		switch {
		case bytes.Equal(ircMsg.Command, []byte{80, 82, 73, 86, 77, 83, 71}): // PRIVMSG
			if c.OnPrivateMessage != nil {
				c.OnPrivateMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{87, 72, 73, 83, 80, 69, 82}): // WHISPER
			if c.OnWhisperMessage != nil {
				c.OnWhisperMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{48, 48, 49}): // RPL_WELCOME (001) Welcome, GLHF!
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

			if c.OnConnect != nil {
				c.OnConnect(true)
			}

		case bytes.Equal(ircMsg.Command, []byte{48, 48, 50}): // RPL_YOURHOST (002) Your host is tmi.twitch.tv
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{48, 48, 51}): // RPL_CREATED (003) This server is rather new
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{48, 48, 52}): // RPL_MYINFO (004)
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{51, 53, 51}): // RPL_NAMREPLY (353)
			if c.OnNamesMessage != nil {
				c.OnNamesMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{51, 54, 54}): // RPL_ENDOFNAMES (366)
			if c.OnEndOfNamesMessage != nil {
				c.OnEndOfNamesMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{51, 55, 50}): // RPL_MOTD (372) You are in a maze of twisty passages, all alike.
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{51, 55, 53}): // RPL_MOTDSTART (375)
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{51, 55, 54}): // RPL_ENDOFMOTD (376)
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{67, 65, 80}): // CAP
			if c.Debug {
				log.Printf(debugTemplate, v)
			}

		case bytes.Equal(ircMsg.Command, []byte{72, 79, 83, 84, 84, 65, 82, 71, 69, 84}): // HOSTTARGET
			if c.OnHosttargetMessage != nil {
				c.OnHosttargetMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{78, 79, 84, 73, 67, 69}): // NOTICE
			if c.OnNoticeMessage != nil {
				c.OnNoticeMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{71, 76, 79, 66, 65, 76, 85, 83, 69, 82, 83, 84, 65, 84, 69}): // GLOBALUSERSTATE
			if c.OnGlobalUserSateMessage != nil {
				c.OnGlobalUserSateMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 83, 84, 65, 84, 69}): // USERSTATE
			if c.OnUserStateMessage != nil {
				c.OnUserStateMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{82, 79, 79, 77, 83, 84, 65, 84, 69}): // ROOMSTATE
			if c.OnRoomStateMessage != nil {
				c.OnRoomStateMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 77, 83, 71}): // CLEARMSG
			if c.OnClearMsgMessage != nil {
				c.OnClearMsgMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 67, 72, 65, 84}): // CLEARCHAT
			if c.OnClearChatMessage != nil {
				c.OnClearChatMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 78, 79, 84, 73, 67, 69}): // USERNOTICE
			if c.OnUserNoticeMessage != nil {
				c.OnUserNoticeMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{80, 73, 78, 71}): // PING // https://blog.golang.org/concurrency-timeouts
			if c.IsConnected() {
				c.write([]byte{80, 79, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PONG :tmi.twitch.tv\r\n"
			}

		case bytes.Equal(ircMsg.Command, []byte{80, 79, 78, 71}): // PONG
			c.pongReceived <- true

		case bytes.Equal(ircMsg.Command, []byte{74, 79, 73, 78}): // JOIN
			if c.OnJoinMessage != nil {
				c.OnJoinMessage(*ircMsg)
			}

		case bytes.Equal(ircMsg.Command, []byte{80, 65, 82, 84}): // PART
			if c.OnPartMessage != nil {
				c.OnPartMessage(*ircMsg)
			}

		default:
			if c.OnUnknownMessage != nil {
				c.OnUnknownMessage(*ircMsg)
			}
		}

	}
}

func (c *Client) sendJoin(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_joinRateQueueLimit.add()
		go func() {
			time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			_joinRateQueueLimit.sub()
		}()

		if c.BotVerified {
			if _joinRateQueueLimit.get() >= verifiedJoinRateLimitMessages {
				log.Printf(joinRateQueueLimitTemplate, verifiedJoinRateLimitMessages)
				time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			}
		} else {
			if _joinRateQueueLimit.get() >= joinRateLimitMessages {
				log.Printf(joinRateQueueLimitTemplate, joinRateLimitMessages)
				time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			}
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendAuthenticate(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_authenticateRateQueueLimit.add()

		go func() {
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			_authenticateRateQueueLimit.sub()
		}()

		if c.BotVerified {
			if _authenticateRateQueueLimit.get() >= verifiedauthenticateRateLimitMessages {
				log.Printf(authenticateRateQueueLimitTemplate, verifiedauthenticateRateLimitMessages)
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		} else {
			if _authenticateRateQueueLimit.get() >= authenticateRateLimitMessages {
				log.Printf(authenticateRateQueueLimitTemplate, authenticateRateLimitMessages)
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) send(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimit.add()
		go func() {
			time.Sleep(time.Duration(rateLimitSeconds) * time.Second)
			_queueRateLimit.sub()
		}()

		if _queueRateLimit.get() >= rateLimitMessages {
			log.Printf(queueRateLimitTemplate, rateLimitMessages)
			time.Sleep(time.Duration(rateLimitSeconds) * time.Second)
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendModOp(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitModOp.add()
		go func() {
			time.Sleep(time.Duration(rateLimitModOpSeconds) * time.Second)
			_queueRateLimitModOp.sub()
		}()

		if _queueRateLimitModOp.get() >= rateLimitModOpMessages {
			log.Printf(queueRateLimitModOpTemplate, rateLimitModOpMessages)
			time.Sleep(time.Duration(rateLimitModOpSeconds) * time.Second)
		}
		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendWhisper(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitWhisperMinute.add()
		_queueRateLimitWhisperSeconds.add()
		go func() {
			time.Sleep(time.Duration(rateLimitWhisperMinute) * time.Minute)
			_queueRateLimitWhisperMinute.sub()
		}()
		go func() {
			time.Sleep(time.Duration(rateLimitWhisperSeconds) * time.Second)
			_queueRateLimitWhisperSeconds.sub()
		}()

		if _queueRateLimitWhisperMinute.get() >= rateLimitWhisperMinuteMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitWhisperMinuteMessages)
			time.Sleep(time.Duration(rateLimitWhisperMinute) * time.Minute)
		}
		if _queueRateLimitWhisperSeconds.get() >= rateLimitWhisperSecondsMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitWhisperSecondsMessages)
			time.Sleep(time.Duration(rateLimitWhisperSeconds) * time.Second)
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendWhisperKnownBot(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitKnownBotsWhisperMinute.add()
		_queueRateLimitKnownBotsWhisperSeconds.add()
		go func() {
			time.Sleep(time.Duration(rateLimitKnownBotsWhisperMinute) * time.Minute)
			_queueRateLimitKnownBotsWhisperMinute.sub()
		}()
		go func() {
			time.Sleep(time.Duration(rateLimitKnownBotsWhisperSeconds) * time.Second)
			_queueRateLimitKnownBotsWhisperSeconds.sub()
		}()

		if _queueRateLimitKnownBotsWhisperMinute.get() >= rateLimitKnownBotsWhisperMinuteMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitKnownBotsWhisperMinuteMessages)
			time.Sleep(time.Duration(rateLimitKnownBotsWhisperMinute) * time.Minute)
		}
		if _queueRateLimitKnownBotsWhisperSeconds.get() >= rateLimitKnownBotsWhisperSecondsMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitKnownBotsWhisperSecondsMessages)
			time.Sleep(time.Duration(rateLimitKnownBotsWhisperSeconds) * time.Second)
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendWhisperVerifiedBots(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitVerifiedBotsWhisperMinute.add()
		_queueRateLimitVerifiedBotsWhisperSeconds.add()
		go func() {
			time.Sleep(time.Duration(rateLimitVerifiedBotsWhisperMinute) * time.Minute)
			_queueRateLimitVerifiedBotsWhisperMinute.sub()
		}()
		go func() {
			time.Sleep(time.Duration(rateLimitVerifiedBotsWhisperSeconds) * time.Second)
			_queueRateLimitVerifiedBotsWhisperSeconds.sub()
		}()

		if _queueRateLimitVerifiedBotsWhisperMinute.get() >= rateLimitVerifiedBotsWhisperMinuteMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitVerifiedBotsWhisperMinuteMessages)
			time.Sleep(time.Duration(rateLimitVerifiedBotsWhisperMinute) * time.Minute)
		}
		if _queueRateLimitVerifiedBotsWhisperSeconds.get() >= rateLimitVerifiedBotsWhisperSecondsMessages {
			log.Printf(queueRateLimitWhisperTemplate, rateLimitVerifiedBotsWhisperSecondsMessages)
			time.Sleep(time.Duration(rateLimitVerifiedBotsWhisperSeconds) * time.Second)
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) pingPong() { // https://github.com/gempir/go-twitch-irc/blob/f5ac4c45474ea2fb0e5f1f77f0bd7bbbcc70da7c/c.go#L791
	c.pongReceived = make(chan bool, 1)
	var pingTime time.Time
	if c.IsConnected() {
		c.write([]byte{80, 73, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PING :tmi.twitch.tv\r\n"
	}
	go func() {
		for {
			// About once every five minutes, the server will send you a PING :tmi.twitch.tv. To ensure that your connection to the server is not prematurely terminated, reply with PONG :tmi.twitch.tv.
			<-time.After(3 * 60 * time.Second)
			if c.IsConnected() {
				c.write([]byte{80, 73, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PING :tmi.twitch.tv\r\n"
				if c.OnPongLatency != nil {
					pingTime = time.Now()
				}
			}
			select {
			case <-c.context.Done():
				return
			case <-c.pongReceived:
				// Received pong message within the time limit, we're good
				if c.OnPongLatency != nil {
					pongTime := time.Now()
					c.OnPongLatency(pongTime.Sub(pingTime))
				}
				continue

			case <-time.After(time.Second * 5):
				log.Println("// No pong message was received within the pong timeout, reconnect")
				c.CloseAndReconnect()
			}
		}
	}()
}
