package websocket

import (
	"github.com/gorilla/websocket"
	"log"
)

type ClientList map[*Client]bool
type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress (means outgoing data) is used to avoid concurrent writes on the websocket connection
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

func (c *Client) readMessages() {
	defer func() {
		//cleanup connection
		c.manager.removeClient(c)
	}()

	//continuously read messages from the connection
	for {
		//message types: ping, pong, data, binary
		messageType, payload, err := c.connection.ReadMessage()
		if err != nil {
			//check for abnormal closes to connection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected close error:", err)
			}
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Println("Normal close error:", err)
			}
			break
		}

		log.Println(messageType)
		log.Println(string(payload))
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		// <- Data flows from the channel "c" into value "message"
		case message, ok := <-c.egress:
			if !ok {
				return
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error writing message:", err)
				return
			}
		}
	}
}
