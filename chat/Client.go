package chat

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Client struct {
	Name  string `json:"name"`
	conn  *websocket.Conn
	ws    *WsServer
	send  chan []byte
	rooms map[*Room]bool
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

func NewClient(conn *websocket.Conn, ws *WsServer, name string) *Client {
	return &Client{
		Name:  name,
		conn:  conn,
		ws:    ws,
		send:  make(chan []byte),
		rooms: make(map[*Room]bool),
	}
}

func (c *Client) getName() string {
	return c.Name
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
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		c.handleNewMessage(jsonMessage)
	}

}

func (c *Client) isInRoom(room *Room) bool {
	for _, ok := c.rooms[room]; ok; {
		return true
	}
	return false
}

func (c *Client) joinRoom(message *Message, sender *Client) {

	room := c.ws.findRoomByName(message.Message)

	if room == nil {
		room = c.ws.createRoom(message.Message, sender != nil)
	}

	if sender == nil && room.Private {
		return
	}

	if !c.isInRoom(room) {
		c.rooms[room] = true
		room.register <- c

		c.notifyRoomOnJoin(room, message)
	}
}

func (c *Client) notifyRoomOnJoin(room *Room, message *Message) {

	msg := &Message{
		Action:  RoomJoinedAction,
		Message: c.getName() + " joined the room!",
		Sender:  message.Sender,
	}

	room.broadcast <- msg

}

func (c *Client) sendMessageAction(message *Message) {

	room := c.ws.findRoomByName(message.Target)
	if room != nil {
		room.broadcast <- message
	}

}

func (c *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	err := json.Unmarshal(jsonMessage, &message)
	handleError2(err)

	// attaching the client to the sender of the message
	message.Sender = c

	switch message.Action {
	case SendMessageAction:
		c.sendMessageAction(&message)

	case JoinRoomAction:
		c.joinRoom(&message, message.Sender)

	case LeaveRoomAction:
		room := c.ws.findRoomByName(message.Target)
		if room != nil {
			room.logout <- c
		}
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
		case message, ok := <-c.send:
			log.Println(message)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			handleError2(err)

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
	for room := range c.rooms {
		room.logout <- c
	}
	close(c.send)
	c.conn.Close()
}
