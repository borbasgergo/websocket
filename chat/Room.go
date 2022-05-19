package chat

import (
	"github.com/google/uuid"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const RoomJoinedAction = "joined-to-room"

type Room struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	clients   map[*Client]bool
	register  chan *Client
	logout    chan *Client
	broadcast chan *Message
	Private   bool `json:"private"`
}

func NewRoom(name string, isPrivate bool) *Room {
	return &Room{
		Name:      name,
		clients:   make(map[*Client]bool),
		register:  make(chan *Client),
		logout:    make(chan *Client),
		broadcast: make(chan *Message),
		Private:   isPrivate,
	}
}

func (room *Room) getName() string {
	return room.Name
}

func (room *Room) startListening() {

	for {
		select {
		case client := <-room.register:
			room.registerClient(client)
		case client := <-room.logout:
			room.logoutClient(client)
		case message := <-room.broadcast:
			log.Println(message)
			room.broadcastToRoom(message.encode())
		}
	}

}

func (room *Room) registerClient(client *Client) {
	room.clients[client] = true
}

func (room *Room) logoutClient(client *Client) {
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}
}

func (room *Room) broadcastToRoom(message []byte) {
	log.Println(room.clients)
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) getId() string {
	return room.Id.String()
}
