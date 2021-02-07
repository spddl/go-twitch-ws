// +build windows linux js,wasm

// Package recws provides websocket client based on gorilla/websocket
// that will automatically reconnect if the connection is dropped.
// stolen from https://github.com/recws-org/recws

package twitch

import (
	"log"
	"time"

	"nhooyr.io/websocket"
)

// CloseAndReconnect will try to reconnect.
func (c *Client) CloseAndReconnect() {
	if c.getConn() != nil {
		c.mu.Lock()
		c.conn.Close(websocket.StatusNormalClosure, "")
		c.mu.Unlock()
	}
	c.mu.Lock()
	c.isConnected = false
	c.mu.Unlock()
	go c.connect()
}

func (c *Client) getConn() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn
}

// Close closes the underlying network connection without
// sending or waiting for a close frame.
func (c *Client) Close() {
	if c.getConn() != nil {
		c.mu.Lock()
		c.cancel()
		c.conn.Close(websocket.StatusNormalClosure, "")
		c.mu.Unlock()
	}

	c.mu.Lock()
	select {
	case <-c.emitQueue.Authenticate:
	default:
		close(c.emitQueue.Authenticate)
	}

	select {
	case <-c.emitQueue.Join:
	default:
		close(c.emitQueue.Join)
	}

	select {
	case <-c.emitQueue.RateLimit:
	default:
		close(c.emitQueue.RateLimit)
	}

	select {
	case <-c.emitQueue.ModOp:
	default:
		close(c.emitQueue.ModOp)
	}

	select {
	case <-c.emitQueue.Whisper:
	default:
		close(c.emitQueue.Whisper)
	}

	c.isConnected = false
	c.mu.Unlock()
}

func (c *Client) Run() {
	// Connect
	go c.connect()

	// wait on first attempt
	time.Sleep(2 * time.Second)
}

func (c *Client) connect() {
	delay := 0
	for {
		select {
		case <-c.context.Done():
			return
		default:
			wsConn, _, err := websocket.Dial(c.context, c.Server, nil)
			c.mu.Lock()
			c.conn = wsConn
			c.isConnected = err == nil
			c.mu.Unlock()

			if err != nil {
				log.Println(err)
			} else {
				if c.Debug {
					log.Printf("Connection was successfully established with %s\n", c.Server)
				}
				c.OnConnect = func(status bool) {
					c.joinCommand(c.Channel)
				}
				c.login()
				return
			}

			if delay != 0 {
				if c.Debug {
					log.Println("Reconnect: will try again in", delay, "seconds.")
				}
				time.Sleep(time.Duration(delay) * time.Second)
				delay *= 2
				if delay > 600 { // 10 min
					delay = 600
				}
			} else {
				delay = 1
			}
		}
	}
}

// IsConnected returns the WebSocket connection state
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isConnected
}
