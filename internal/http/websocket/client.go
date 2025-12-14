package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type ClientList map[*Client]bool
type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress (means outgoing data) is used to avoid concurrent writes on the websocket connection
	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
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
		_, payload, err := c.connection.ReadMessage()
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

		var request Event

		//unmarshal json Event message into request
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handling Event message:", err)
			continue
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		//When messages are sent, we will write them to egress channel, which will one by one, select them and fire them onto the websocket connection
		// <- Data flows from the channel "c" into value "message"
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Error writing message; connection closed:", err)
					return
				}
			}

			//marshal message into a JSON
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("Error marshalling message:", err)
				return
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Error writing message:", err)
			}
			log.Println("message sent")

		}
	}
}
