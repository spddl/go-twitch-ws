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

func NewClient(client Client) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client.ctx = &ctx
	client.cancel = cancel

	c, _, err := websocket.Dial(ctx, "wss://irc-ws.chat.twitch.tv", nil)
	if err != nil {
		log.Fatal(err)
	}
	client.conn = c

	if client.Debug {
		log.Printf("< connecting to %s\n", client.Server)
	}

	if client.User == "" {
		client.User = fmt.Sprintf("justinfan%d", rand.Intn(9999-1000)+1000)
	}

	client.emitQueue.Authenticate = make(chan string)
	client.emitQueue.Join = make(chan string)
	client.emitQueue.RateLimit = make(chan string)
	client.emitQueue.ModOp = make(chan string)

	go client.sendAuthenticate(client.emitQueue.Authenticate)
	go client.sendJoin(client.emitQueue.Join)
	go client.send(client.emitQueue.RateLimit)
	go client.sendModOp(client.emitQueue.ModOp)

	go client.pingPong() // takes care of the ping pong

	go client.read()

	return &client, nil
}

func (client *Client) Close() {
	if client.Debug {
		log.Println("Close Connection")
	}

	close(client.emitQueue.Authenticate)
	close(client.emitQueue.Join)
	close(client.emitQueue.RateLimit)
	close(client.emitQueue.ModOp)

	err := client.conn.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		panic(err)
	}

	client.cancel()
}
