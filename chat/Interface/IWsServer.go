package Interface

import (
	"websocketServer/chat"
)

type IWsServer interface {
	Run()
	RegisterClient(client *chat.Client)
	LogoutClient(client *chat.Client)
}
