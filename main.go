package main

import (
	"log"
	"net/http"
	"websocketServer/chat"
)

func main() {

	webServer := chat.NewWsServer()
	go webServer.Run()

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		chat.WSHandler(*webServer, writer, request)
	})

	log.Fatal(http.ListenAndServe(":8001", nil))
}
