package twitch

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"nhooyr.io/websocket"
)

type EmitQueue struct {
	Authenticate chan string
	Join         chan string
	RateLimit    chan string
	ModOp        chan string
	Whisper      chan string
}

type Client struct {
	Server string
	User   string
	Oauth  string
	Debug  bool

	conn   *websocket.Conn
	ctx    *context.Context
	cancel context.CancelFunc

	BotVerified bool

	IsConnected bool

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
	OnPongLatency           func(message time.Duration)
}

func NewClient(c Client) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = &ctx
	c.cancel = cancel

	conn, _, err := websocket.Dial(ctx, c.Server, nil)
	if err != nil {
		log.Fatal(err)
	}
	c.conn = conn

	if c.Debug {
		log.Printf("< connecting to %s\n", c.Server)
	}

	if c.User == "" {
		c.User = fmt.Sprintf("justinfan%d", rand.Intn(9999-1000)+1000)
	}

	c.emitQueue.Authenticate = make(chan string)
	c.emitQueue.Join = make(chan string)
	c.emitQueue.RateLimit = make(chan string)
	c.emitQueue.ModOp = make(chan string)

	go c.sendAuthenticate(c.emitQueue.Authenticate)
	go c.sendJoin(c.emitQueue.Join)
	go c.send(c.emitQueue.RateLimit)
	go c.sendModOp(c.emitQueue.ModOp)
	go c.sendWhisper(c.emitQueue.Whisper)

	go c.pingPong() // takes care of the ping pong

	go c.read()

	return &c, nil
}

func (c *Client) Close() {
	if c.Debug {
		log.Println("Close Connection")
	}

	close(c.emitQueue.Authenticate)
	close(c.emitQueue.Join)
	close(c.emitQueue.RateLimit)
	close(c.emitQueue.ModOp)

	c.cancel()
	// err := c.conn.Close(websocket.StatusNormalClosure, "")
	// if err != nil {
	// 	panic(err)
	// }
}
