package chat

import (
	"errors"
	"log"
)

type WsServer struct {
	Clients   map[*Client]bool
	Register  chan *Client
	Logout    chan *Client
	Broadcast chan []byte
	Rooms     map[*Room]bool
}

func NewWsServer() *WsServer {
	return &WsServer{
		Clients:   make(map[*Client]bool),
		Register:  make(chan *Client),
		Logout:    make(chan *Client),
		Broadcast: make(chan []byte),
		Rooms:     make(map[*Room]bool),
	}
}

func (ws *WsServer) Run() {
	for {
		select {
		case client := <-ws.Register:
			ws.RegisterClient(client)

		case client := <-ws.Logout:
			ws.LogoutClient(client)

		case message := <-ws.Broadcast:
			ws.BroadcastMessage(message)
		}
	}
}

func (ws *WsServer) RegisterClient(client *Client) {
	ws.Clients[client] = true
}

func (ws *WsServer) LogoutClient(client *Client) {
	if _, ok := ws.Clients[client]; ok {
		delete(ws.Clients, client)
	}
}

func (ws *WsServer) BroadcastMessage(message []byte) {
	for client := range ws.Clients {
		client.send <- message
	}
}

func (ws *WsServer) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range ws.Rooms {
		if room.getName() == name {
			foundRoom = room
			return foundRoom
		}
	}

	return foundRoom
}
func (ws *WsServer) findRoomById(id string) (room *Room, err error) {
	for room := range ws.Rooms {
		if room.getId() == id {
			return room, nil
		}
	}
	return nil, errors.New("error")
}

func (ws *WsServer) createRoom(name string, isPrivate bool) *Room {
	room := NewRoom(name, isPrivate)
	log.Println(room)
	go room.startListening()
	ws.Rooms[room] = true

	return room
}
