// +build windows js,wasm

package twitch

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type EmitQueue struct {
	Authenticate        chan string
	Join                chan string
	RateLimit           chan string
	ModOp               chan string
	Whisper             chan string
	WhisperKnownBot     chan string
	WhisperVerifiedBots chan string
}

type Client struct {
	Server      string
	User        string
	Oauth       string
	Debug       bool
	BotVerified bool
	BotKnown    bool
	Channel     []string

	conn    *websocket.Conn
	context context.Context
	cancel  context.CancelFunc

	isConnected bool
	mu          sync.RWMutex

	emitQueue    EmitQueue
	pongReceived chan bool

	OnConnect               func(message bool)
	OnPrivateMessage        func(message IRCMessage)
	OnRoomStateMessage      func(message IRCMessage)
	OnHosttargetMessage     func(message IRCMessage)
	OnNoticeMessage         func(message IRCMessage)
	OnJoinMessage           func(message IRCMessage)
	OnPartMessage           func(message IRCMessage)
	OnUnknownMessage        func(message IRCMessage)
	OnUserNoticeMessage     func(message IRCMessage)
	OnClearMsgMessage       func(message IRCMessage)
	OnClearChatMessage      func(message IRCMessage)
	OnGlobalUserSateMessage func(message IRCMessage)
	OnUserStateMessage      func(message IRCMessage)
	OnNamesMessage          func(message IRCMessage)
	OnEndOfNamesMessage     func(message IRCMessage)
	OnWhisperMessage        func(message IRCMessage)
	OnPongLatency           func(message time.Duration)
}

func NewClient(c *Client) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c.mu.Lock()
	c.context = ctx
	c.cancel = cancel
	c.emitQueue.Authenticate = make(chan string)
	c.emitQueue.Join = make(chan string)
	c.emitQueue.RateLimit = make(chan string)
	c.emitQueue.ModOp = make(chan string)
	c.emitQueue.Whisper = make(chan string)
	c.emitQueue.WhisperKnownBot = make(chan string)
	c.emitQueue.WhisperVerifiedBots = make(chan string)
	if c.User == "" {
		c.User = fmt.Sprintf("justinfan%d", rand.Intn(9999-1000)+1000)
	}
	c.mu.Unlock()

	go c.sendAuthenticate(c.emitQueue.Authenticate)
	go c.sendJoin(c.emitQueue.Join)
	go c.send(c.emitQueue.RateLimit)
	go c.sendModOp(c.emitQueue.ModOp)
	go c.sendWhisper(c.emitQueue.Whisper)

	go c.pingPong() // takes care of the ping pong

	go c.read()
	return c, nil
}
