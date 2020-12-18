// +build windows

package twitch

import (
	"bytes"
)

func (c *Client) read() {
	for {
		select {
		case <-c.context.Done():
			c.Close()
			return
		default:
			if c.IsConnected() {
				_, r, err := c.conn.Reader(c.context)
				if err != nil {
					c.CloseAndReconnect()
					continue
				}

				var buf bytes.Buffer
				buf.Grow(bytes.MinRead)

				_, err = buf.ReadFrom(r)
				if err != nil {
					continue // failed to read: WebSocket closed: sent close frame: status = StatusNormalClosure and reason = ""
				}

				msg := bytes.Split(buf.Bytes(), []byte{13, 10}) // "\r\n"
				for _, value := range msg {
					go c.parser(value)
				}
			}
		}
	}
}
