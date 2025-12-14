package websocket

import (
	"fmt"
	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/http/handlers"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type Manager struct {
	App          *app.App
	User         *handlers.UserHandler
	Post         *handlers.PostHandler
	Comment      *handlers.CommentHandler
	Reaction     *handlers.ReactionHandler
	Channel      *handlers.ChannelHandler
	Mod          *handlers.ModHandler
	Clients      ClientList
	sync.RWMutex //read/write lock in Go. It protects shared data when multiple goroutines access it, allowing many readers at the same time but only one writer at a time.
	// Notification *NotificationHandler
	// Membership *MembershipHandler
}

func NewManager() *Manager {
	return &Manager{
		Clients: make(ClientList), //creates a client list whenever a new manager is created so no nil pointer exception
	}
}

func (ws *Manager) ServeWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Upgrade the HTTP connection to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := NewClient(conn, ws)

	ws.addClient(client)

	//Start client processes
	go client.readMessages()
	go client.writeMessages()
}

func (ws *Manager) addClient(client *Client) {
	//when 2 people connecting at the same time, the map won't get motified at the same time
	ws.Lock()

	//it will unlock once the map is modified
	defer ws.Unlock()

	//whenever a new client is added, add bool that it's connected
	ws.Clients[client] = true
}

func (ws *Manager) removeClient(client *Client) {
	ws.Lock()
	defer ws.Unlock()

	if _, ok := ws.Clients[client]; ok {
		if err := client.connection.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
		delete(ws.Clients, client)
	}
}
