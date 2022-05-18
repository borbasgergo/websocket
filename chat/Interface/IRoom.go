package Interface

import "websocketServer/chat"

type IRoom interface {
	registerClient(c *chat.Client)
	logoutClient(c *chat.Client)
	broadcastToRoom(message []byte)
	getName() string
}
