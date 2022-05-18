package chat

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func handleError(msg string, err error) {
	if err != nil {
		log.Fatal(msg + ":" + err.Error())
	}
}

func handleUpgrade(
	w http.ResponseWriter,
	r *http.Request) *websocket.Conn {

	conn, err := upgrader.Upgrade(w, r, nil)
	handleError("Connection error", err)

	return conn
}

func WSHandler(
	server WsServer,
	w http.ResponseWriter,
	r *http.Request) {

	conn := handleUpgrade(w, r)

	clientName := r.URL.Query()["name"]

	client := NewClient(conn, &server, clientName[0])

	go client.read()
	go client.write()

	server.Register <- client

}
