package websocket

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"` //not marshalled, any type of event messsage can be sent
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
)

type SendMessageEvent struct {
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

type NewMessageEvent struct {
	ChatID    string `json:"chat_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
	Sender    struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	} `json:"sender"`
	Created time.Time `json:"created"`
}
