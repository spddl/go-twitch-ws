package twitch

import (
	"bytes"
	"log"
	"time"
)

func (c *Client) read() {
	for {
		_, r, err := c.conn.Reader(*c.ctx)
		if err != nil {
			log.Println(err)
			// messageQueue.go:12: failed to get reader: WebSocket closed: sent close frame: status = StatusNormalClosure and reason = ""
			// messageQueue.go:12: failed to get reader: failed to acquire lock: context canceled
			return
		}

		var buf bytes.Buffer
		buf.Grow(bytes.MinRead)

		_, err = buf.ReadFrom(r)
		if err != nil {
			log.Println(err)
		}

		msg := bytes.Split(buf.Bytes(), []byte{13, 10}) // "\r\n"
		for _, value := range msg {
			go c.parser(value)
		}
	}
}

func (c *Client) write(msg []byte) {
	w, err := c.conn.Writer(*c.ctx, 1) // websocket.MessageText
	if err != nil {
		log.Println(err)
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
		log.Printf("< %s", msg)
	}
}

func (c *Client) parser(msgData []byte) {
	msg := bytes.Split(msgData, []byte{13, 10}) // "\r\n"
	for _, v := range msg {
		if len(v) == 0 {
			continue
		}

		ircMsg, err := parseIRCMessage(v)
		if err != nil {
			log.Println("parseIRCMessage:", err)
			return
		}

		if bytes.Equal(ircMsg.Command, []byte{80, 82, 73, 86, 77, 83, 71}) { // PRIVMSG
			if c.OnPrivateMessage != nil {
				c.OnPrivateMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 49}) { // RPL_WELCOME (001) Welcome, GLHF!
			if c.OnConnect != nil {
				c.OnConnect(true)
			}
			if c.Debug {
				log.Printf("> %s", v)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 50}) { // RPL_YOURHOST (002) Your host is tmi.twitch.tv
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 51}) { // RPL_CREATED (003) This server is rather new
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 52}) { // RPL_MYINFO (004)
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 53, 51}) { // RPL_NAMREPLY (353)
			if c.OnNamesMessage != nil {
				c.OnNamesMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 54, 54}) { // RPL_ENDOFNAMES (366)
			if c.OnEndOfNamesMessage != nil {
				c.OnEndOfNamesMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 50}) { // RPL_MOTD (372) You are in a maze of twisty passages, all alike.
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 53}) { // RPL_MOTDSTART (375)
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 54}) { // RPL_ENDOFMOTD (376)
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{67, 65, 80}) { // CAP
			if c.Debug {
				log.Printf("> %s", v)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{72, 79, 83, 84, 84, 65, 82, 71, 69, 84}) { // HOSTTARGET
			if c.OnHosttargetMessage != nil {
				c.OnHosttargetMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{78, 79, 84, 73, 67, 69}) { // NOTICE
			if c.OnNoticeMessage != nil {
				c.OnNoticeMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{71, 76, 79, 66, 65, 76, 85, 83, 69, 82, 83, 84, 65, 84, 69}) { // GLOBALUSERSTATE
			if c.OnGlobalUserSateMessage != nil {
				c.OnGlobalUserSateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 83, 84, 65, 84, 69}) { // USERSTATE
			if c.OnUserStateMessage != nil {
				c.OnUserStateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{82, 79, 79, 77, 83, 84, 65, 84, 69}) { // ROOMSTATE
			if c.OnRoomStateMessage != nil {
				c.OnRoomStateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 77, 83, 71}) { // CLEARMSG
			if c.OnClearMsgMessage != nil {
				c.OnClearMsgMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 67, 72, 65, 84}) { // CLEARCHAT
			if c.OnClearChatMessage != nil {
				c.OnClearChatMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 78, 79, 84, 73, 67, 69}) { // USERNOTICE
			if c.OnUserNoticeMessage != nil {
				c.OnUserNoticeMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{80, 73, 78, 71}) { // PING // https://blog.golang.org/concurrency-timeouts
			c.write([]byte{80, 79, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PONG :tmi.twitch.tv\r\n"

		} else if bytes.Equal(ircMsg.Command, []byte{80, 79, 78, 71}) { // PONG
			c.pongReceived <- true

		} else if bytes.Equal(ircMsg.Command, []byte{74, 79, 73, 78}) { // JOIN
			if c.OnJoinMessage != nil {
				c.OnJoinMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{80, 65, 82, 84}) { // PART
			if c.OnPartMessage != nil {
				c.OnPartMessage(*ircMsg)
			}

		} else {
			if c.OnUnknownMessage != nil {
				c.OnUnknownMessage(*ircMsg)
			}

		}
	}
}

func (c *Client) sendJoin(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_joinRateQueueLimit.Add()
		go func() {
			time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			_joinRateQueueLimit.Sub()
		}()

		if c.BotVerified {
			if _joinRateQueueLimit.Get() >= verifiedJoinRateLimitMessages {
				log.Printf("_joinRateQueueLimit(%d) limit reached", verifiedJoinRateLimitMessages)
				time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			}
		} else {
			if _joinRateQueueLimit.Get() >= joinRateLimitMessages {
				log.Printf("_joinRateQueueLimit(%d) limit reached", joinRateLimitMessages)
				time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			}
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendAuthenticate(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_authenticateRateQueueLimit.Add()

		go func() {
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			_authenticateRateQueueLimit.Sub()
		}()

		if c.BotVerified {
			if _authenticateRateQueueLimit.Get() >= verifiedauthenticateRateLimitMessages {
				log.Printf("_authenticateRateQueueLimit(%d) limit reached", verifiedauthenticateRateLimitMessages)
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		} else {
			if _authenticateRateQueueLimit.Get() >= authenticateRateLimitMessages {
				log.Printf("_authenticateRateQueueLimit(%d) limit reached", authenticateRateLimitMessages)
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) send(rawMsgChan <-chan string) {
	// log.Println("send")
	for rawMsg := range rawMsgChan {
		_queueRateLimit.Add()
		go func() {
			time.Sleep(time.Duration(rateLimitSeconds) * time.Second)
			_queueRateLimit.Sub()
		}()

		if _queueRateLimit.Get() >= rateLimitMessages {
			log.Printf("_queueRateLimit(%d) limit reached", rateLimitMessages)
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
		}
		log.Println("send", _queueRateLimit.Get())
		log.Println("rawMsg", rawMsg)
		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendModOp(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitModOp.Add()
		go func() {
			time.Sleep(time.Duration(rateLimitModOpSeconds) * time.Second)
			_queueRateLimitModOp.Sub()
		}()

		if _queueRateLimitModOp.Get() >= rateLimitModOpMessages {
			log.Printf("_queueRateLimitModOp(%d) limit reached", rateLimitModOpMessages)
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
		}

		c.write([]byte(rawMsg))
	}
}

func (c *Client) sendWhisper(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		c.write([]byte(rawMsg))
	}
}

func (c *Client) pingPong() { // https://github.com/gempir/go-twitch-irc/blob/f5ac4c45474ea2fb0e5f1f77f0bd7bbbcc70da7c/c.go#L791
	c.pongReceived = make(chan bool, 1)
	var pingTime time.Time
	c.write([]byte{80, 73, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PING :tmi.twitch.tv\r\n"

	go func() {
		for {
			<-time.After(time.Second * 60)                                                                                // https://github.com/tmijs/tmi.js/blob/4bb66c433b8ae28326b4cd8567357e6ea729e91a/lib/c.js#L191
			c.write([]byte{80, 73, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PING :tmi.twitch.tv\r\n"
			if c.OnPongLatency != nil {
				pingTime = time.Now()
			}

			select {
			case <-c.pongReceived:
				// Received pong message within the time limit, we're good
				if c.OnPongLatency != nil {
					pongTime := time.Now()
					c.OnPongLatency(pongTime.Sub(pingTime))
				}
				continue

			case <-time.After(time.Second * 5):
				// No pong message was received within the pong timeout, disconnect
				// c.cReconnect.Close()
				// closer.Close()
			}
		}
	}()
}
