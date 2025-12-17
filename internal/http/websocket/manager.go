package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/sqlite"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Manager struct {
	Clients       ClientList
	sync.RWMutex  //read/write lock in Go. It protects shared data when multiple goroutines access it, allowing many readers at the same time but only one writer at a time.
	EventHandlers map[string]EventHandler
	OTPs          RetentionMap
	Chats         *sqlite.ChatModel
	Users         *sqlite.UserModel
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		Clients:       make(ClientList), //creates a client list whenever a new manager is created so no nil pointer exception
		EventHandlers: make(map[string]EventHandler),
		OTPs:          NewRetentionMap(ctx, 5*time.Second),
	}
	m.setupEventHandlers()
	return m
}

func (ws *Manager) ServeWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	log.Println("checking OTP")
	//block websocket connection if no OTP is present
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Println("OTP exists")

	log.Println("Checking if OTP is valid")
	//block websocket connection if no OTP is not valid
	otpObj, valid := ws.OTPs.VerifyOTP(otp)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Println("OTP is valid; proceeding to open websocket")

	// Upgrade the HTTP connection to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := NewClient(conn, ws, otpObj.UserID)

	ws.addClient(client)

	//Start client processes
	go client.readMessages()
	go client.writeMessages()
}

func (ws *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := ws.EventHandlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("event type not found")
	}

}

func (ws *Manager) setupEventHandlers() {
	ws.EventHandlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	ctx := context.Background()

	// Unmarshal the event payload
	var sendMsgEvent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &sendMsgEvent); err != nil {
		models.LogErrorWithContext(ctx, "Error unmarshalling send message event", err)
		return fmt.Errorf("failed to unmarshal send message event: %w", err)
	}

	// Validate message content
	if sendMsgEvent.Message == "" {
		return errors.New("message content cannot be empty")
	}

	// Parse chatID
	parsedUUID, err := uuid.Parse(sendMsgEvent.ChatID)
	if err != nil {
		models.LogErrorWithContext(ctx, "Invalid chat ID", err)
		return fmt.Errorf("invalid chat ID: %w", err)
	}
	chatID := models.UUIDField{UUID: parsedUUID}

	// Verify sender is in the chat
	isInChat, err := c.manager.Chats.IsUserInChat(ctx, chatID, c.userID)
	if err != nil {
		models.LogErrorWithContext(ctx, "Error checking if user is in chat", err)
		return fmt.Errorf("failed to verify chat membership: %w", err)
	}
	if !isInChat {
		return errors.New("user is not a member of this chat")
	}

	// Save message to database
	messageID, err := c.manager.Chats.CreateChatMessage(ctx, chatID, c.userID, sendMsgEvent.Message)
	if err != nil {
		models.LogErrorWithContext(ctx, "Error saving message to database", err)
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Get sender information
	sender, err := c.manager.Users.GetUserByID(ctx, c.userID)
	if err != nil {
		models.LogErrorWithContext(ctx, "Error fetching sender info", err)
		return fmt.Errorf("failed to get sender info: %w", err)
	}

	// Create NewMessageEvent for broadcasting
	newMsgEvent := NewMessageEvent{
		ChatID:    sendMsgEvent.ChatID,
		MessageID: messageID.String(),
		Content:   sendMsgEvent.Message,
		Created:   time.Now(),
	}
	newMsgEvent.Sender.ID = sender.ID.String()
	newMsgEvent.Sender.Username = sender.Username
	newMsgEvent.Sender.Avatar = sender.Avatar

	// Broadcast to all chat participants
	if err := c.manager.BroadcastToChatParticipants(ctx, chatID, newMsgEvent); err != nil {
		models.LogErrorWithContext(ctx, "Error broadcasting message", err)
		return fmt.Errorf("failed to broadcast message: %w", err)
	}

	models.LogInfoWithContext(ctx, "Message sent successfully in chat %s by user %s", chatID.String(), sender.Username)
	return nil
}

// BroadcastToChatParticipants sends an event to all connected clients who are participants in the chat
func (ws *Manager) BroadcastToChatParticipants(ctx context.Context, chatID models.UUIDField, newMsgEvent NewMessageEvent) error {
	// Get all participant IDs for this chat
	participantIDs, err := ws.Chats.GetChatParticipantIDs(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat participants: %w", err)
	}

	// Marshal the event payload
	payload, err := json.Marshal(newMsgEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal new message event: %w", err)
	}

	// Create the event
	event := Event{
		Type:    EventNewMessage,
		Payload: payload,
	}

	// Lock to safely iterate over clients
	ws.RLock()
	defer ws.RUnlock()

	// Broadcast to all connected clients who are participants
	broadcastCount := 0
	for client := range ws.Clients {
		// Check if this client's userID is in the participant list
		for _, participantID := range participantIDs {
			if client.userID == participantID {
				// Send to this client's egress channel
				select {
				case client.egress <- event:
					broadcastCount++
				default:
					models.LogWarnWithContext(ctx, "Client egress channel full, skipping user %s", client.userID.String())
				}
				break
			}
		}
	}

	models.LogInfoWithContext(ctx, "Broadcast message to %d/%d participants in chat %s", broadcastCount, len(participantIDs), chatID.String())
	return nil
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

// function to check the origin of the websocket connection; for security
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8888":
		return true
	default:
		log.Println("Origin not allowed")
		return false
	}
}
