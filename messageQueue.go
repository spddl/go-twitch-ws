package twitch

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

func (client *Client) read() {
	for {
		_, r, err := client.conn.Reader(*client.ctx)
		if err != nil {
			// log.Println(err)
			// messageQueue.go:12: failed to get reader: WebSocket closed: sent close frame: status = StatusNormalClosure and reason = ""
			// messageQueue.go:12: failed to get reader: failed to acquire lock: context canceled
			return
		}

		b := bpoolGet()

		_, err = b.ReadFrom(r)
		if err != nil {
			log.Println(err)
			return
		}

		client.parser(b.Bytes())

		bpoolPut(b)
	}
}

func (client *Client) write(msg []byte) {
	w, err := client.conn.Writer(*client.ctx, 1) // websocket.MessageText
	if err != nil {
		log.Println(err)
		return
	}

	_, err = w.Write(msg)
	if err != nil {
		log.Println(err)
		return
	}

	err = w.Close()
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("< %s", msg)
}

func (client *Client) parser(msgData []byte) {
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
			if client.OnPrivateMessage != nil {
				client.OnPrivateMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 49}) { // RPL_WELCOME (001) Welcome, GLHF!
			if client.Debug {
				log.Println(">", string(v))
			}

		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 50}) { // RPL_YOURHOST (002) Your host is tmi.twitch.tv
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 51}) { // RPL_CREATED (003) This server is rather new
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{48, 48, 52}) { // RPL_MYINFO (004)
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 53, 51}) { // RPL_NAMREPLY (353)
			if client.OnNamesMessage != nil {
				client.OnNamesMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 54, 54}) { // RPL_ENDOFNAMES (366)
			if client.OnEndOfNamesMessage != nil {
				client.OnEndOfNamesMessage(*ircMsg)
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 50}) { // RPL_MOTD (372) You are in a maze of twisty passages, all alike.
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 53}) { // RPL_MOTDSTART (375)
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{51, 55, 54}) { // RPL_ENDOFMOTD (376)
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{67, 65, 80}) { // CAP
			if client.Debug {
				log.Println(">", string(v))
			}
		} else if bytes.Equal(ircMsg.Command, []byte{72, 79, 83, 84, 84, 65, 82, 71, 69, 84}) { // HOSTTARGET
			if client.OnHosttargetMessage != nil {
				client.OnHosttargetMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{78, 79, 84, 73, 67, 69}) { // NOTICE
			if client.OnNoticeMessage != nil {
				client.OnNoticeMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{71, 76, 79, 66, 65, 76, 85, 83, 69, 82, 83, 84, 65, 84, 69}) { // GLOBALUSERSTATE
			if client.OnGlobalUserSateMessage != nil {
				client.OnGlobalUserSateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 83, 84, 65, 84, 69}) { // USERSTATE
			if client.OnUserStateMessage != nil {
				client.OnUserStateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{82, 79, 79, 77, 83, 84, 65, 84, 69}) { // ROOMSTATE
			if client.OnRoomStateMessage != nil {
				client.OnRoomStateMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 77, 83, 71}) { // CLEARMSG
			if client.OnClearMsgMessage != nil {
				client.OnClearMsgMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{67, 76, 69, 65, 82, 67, 72, 65, 84}) { // CLEARCHAT
			if client.OnClearChatMessage != nil {
				client.OnClearChatMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{85, 83, 69, 82, 78, 79, 84, 73, 67, 69}) { // USERNOTICE
			if client.OnUserNoticeMessage != nil {
				client.OnUserNoticeMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{80, 73, 78, 71}) { // PING // https://blog.golang.org/concurrency-timeouts
			client.write([]byte{80, 79, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PONG :tmi.twitch.tv\r\n"

		} else if bytes.Equal(ircMsg.Command, []byte{80, 79, 78, 71}) { // PONG
			client.pongReceived <- true

		} else if bytes.Equal(ircMsg.Command, []byte{74, 79, 73, 78}) { // JOIN
			if client.OnJoinMessage != nil {
				client.OnJoinMessage(*ircMsg)
			}

		} else if bytes.Equal(ircMsg.Command, []byte{80, 65, 82, 84}) { // PART
			if client.OnPartMessage != nil {
				client.OnPartMessage(*ircMsg)
			}

		} else {
			if client.OnUnknownMessage != nil {
				client.OnUnknownMessage(*ircMsg)
			}

		}
	}
}

func (client *Client) sendJoin(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_joinRateQueueLimit += 1
		go func() {
			time.Sleep(time.Duration(joinRateLimitSeconds) * time.Second)
			_joinRateQueueLimit -= 1
		}()

		if client.BotVerified {
			if _joinRateQueueLimit >= verifiedJoinRateLimitMessages {
				log.Println(fmt.Sprintf("_joinRateQueueLimit(%d) limit reached", verifiedJoinRateLimitMessages))
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		} else {
			if _joinRateQueueLimit >= joinRateLimitMessages {
				log.Println(fmt.Sprintf("_joinRateQueueLimit(%d) limit reached", joinRateLimitMessages))
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		}

		client.write([]byte(rawMsg))
	}
}

func (client *Client) sendAuthenticate(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_authenticateRateQueueLimit += 1
		go func() {
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			_authenticateRateQueueLimit -= 1
		}()

		if client.BotVerified {
			if _authenticateRateQueueLimit >= verifiedauthenticateRateLimitMessages {
				log.Println(fmt.Sprintf("_authenticateRateQueueLimit(%d) limit reached", verifiedauthenticateRateLimitMessages))
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		} else {
			if _authenticateRateQueueLimit >= authenticateRateLimitMessages {
				log.Println(fmt.Sprintf("_authenticateRateQueueLimit(%d) limit reached", authenticateRateLimitMessages))
				time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
			}
		}

		client.write([]byte(rawMsg))
	}
}

func (client *Client) send(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimit += 1
		go func() {
			time.Sleep(time.Duration(rateLimitSeconds) * time.Second)
			_queueRateLimit -= 1
		}()

		if _queueRateLimit >= rateLimitMessages {
			log.Println(fmt.Sprintf("_queueRateLimit(%d) limit reached", rateLimitMessages))
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
		}

		client.write([]byte(rawMsg))
	}
}

func (client *Client) sendModOp(rawMsgChan <-chan string) {
	for rawMsg := range rawMsgChan {
		_queueRateLimitModOp += 1
		go func() {
			time.Sleep(time.Duration(rateLimitModOpSeconds) * time.Second)
			_queueRateLimitModOp -= 1
		}()

		if _queueRateLimitModOp >= rateLimitModOpMessages {
			log.Println(fmt.Sprintf("_queueRateLimitModOp(%d) limit reached", rateLimitModOpMessages))
			time.Sleep(time.Duration(authenticateRateLimitSeconds) * time.Second)
		}

		client.write([]byte(rawMsg))
	}
}

func (client *Client) pingPong() { // https://github.com/gempir/go-twitch-irc/blob/f5ac4c45474ea2fb0e5f1f77f0bd7bbbcc70da7c/client.go#L791
	client.pongReceived = make(chan bool, 1)
	var pingTime time.Time
	go func() {
		for {
			<-time.After(time.Second * 60)                                                                                     // https://github.com/tmijs/tmi.js/blob/4bb66c433b8ae28326b4cd8567357e6ea729e91a/lib/client.js#L191
			client.write([]byte{80, 73, 78, 71, 32, 58, 116, 109, 105, 46, 116, 119, 105, 116, 99, 104, 46, 116, 118, 13, 10}) // "PING :tmi.twitch.tv\r\n"
			if client.OnPongLatency != nil {
				pingTime = time.Now()
			}

			select {
			case <-client.pongReceived:
				// Received pong message within the time limit, we're good
				if client.OnPongLatency != nil {
					pongTime := time.Now()
					client.OnPongLatency(pongTime.Sub(pingTime))
				}
				continue

			case <-time.After(time.Second * 5):
				// No pong message was received within the pong timeout, disconnect
				// c.clientReconnect.Close()
				// closer.Close()
			}
		}

	}()
}
