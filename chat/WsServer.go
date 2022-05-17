package chat

import "log"

type WsServer struct {
	Clients   map[*Client]bool
	Register  chan *Client
	Logout    chan *Client
	Broadcast chan []byte
}

func NewWsServer() *WsServer {
	return &WsServer{
		Clients:   make(map[*Client]bool),
		Register:  make(chan *Client),
		Logout:    make(chan *Client),
		Broadcast: make(chan []byte),
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
			log.Println("message got")
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
		log.Println(client)
		client.Send <- message
	}
}
