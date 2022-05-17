package chat

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Client struct {
	conn *websocket.Conn
	ws   *WsServer
	Send chan []byte
}

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

func NewClient(conn *websocket.Conn, ws *WsServer) *Client {
	return &Client{
		conn: conn,
		ws:   ws,
		Send: make(chan []byte),
	}
}

func (c *Client) read() {
	defer func() {
		c.disconnect()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, jsonMessage, err := c.conn.ReadMessage()
		log.Println(jsonMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		c.ws.Broadcast <- jsonMessage
	}

}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func handleError2(error error) {
	if error != nil {
		log.Println(error.Error())
	}
}
func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(err)
				return
			}
			_, err2 := w.Write(message)
			handleError2(err2)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) disconnect() {
	c.ws.Logout <- c
	close(c.Send)
	c.conn.Close()
}
