package chat

import (
	"encoding/json"
	"log"
)

type Message struct {
	Action  string  `json:"action"`
	Message string  `json:"message"`
	Target  string  `json:"target"`
	Sender  *Client `json:"sender"`
}

func (message *Message) encode() []byte {

	bytes, error := json.Marshal(message)
	if error != nil {
		log.Println(error)
	}

	return bytes
}
