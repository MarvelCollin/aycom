package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket constants for default values when config is not available
const (
	// Buffer sizes
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
	SendBufferSize  = 256

	// Timeouts (in seconds)
	ReadDeadlineTimeout  = 60
	WriteDeadlineTimeout = 10
	PingInterval         = 54

	// Message size limits
	MaxMessageSize = 4096
)

// Create a websocket upgrader based on configuration
func createUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  Config.WebSocket.ReadBufferSize,
		WriteBufferSize: Config.WebSocket.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development; restrict in production
		},
	}
}

var (
	wsManager     *WebSocketManager
	wsManagerOnce sync.Once
	upgrader      websocket.Upgrader // This will be initialized in InitWebsocketServices
)

// InitWebsocketServices initializes all service WebSocket clients
func InitWebsocketServices() {
	log.Println("Initializing WebSocket service clients...")

	// Initialize the WebSocket manager
	wsManager := GetWebSocketManager()
	if wsManager == nil {
		log.Println("Creating new WebSocket manager")
		// Initialize a new WebSocket manager when needed
	}

	// Initialize the upgrader with the configuration
	upgrader = createUpgrader()

	log.Println("WebSocket service clients initialized successfully")
}

// WebSocketManager manages WebSocket connections and message broadcasting
type WebSocketManager struct {
	clients      map[string]*Client
	chatRooms    map[string]map[string]bool // chatID -> map of clientIDs
	userToClient map[string]string          // userID -> clientID
	register     chan *Client
	unregister   chan *Client
	broadcast    chan BroadcastMessage
	mutex        sync.RWMutex
}

// Client represents a connected WebSocket client
type Client struct {
	ID         string
	UserID     string
	Connection *websocket.Conn
	ChatID     string
	Send       chan []byte
	Manager    *WebSocketManager
}

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	ChatID  string
	Message []byte
	UserID  string
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	manager := &WebSocketManager{
		clients:      make(map[string]*Client),
		chatRooms:    make(map[string]map[string]bool),
		userToClient: make(map[string]string),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan BroadcastMessage),
	}

	go manager.run()
	return manager
}

// run starts the WebSocket manager
func (manager *WebSocketManager) run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client.ID] = client
			manager.userToClient[client.UserID] = client.ID

			// Add to chat room if chatID is provided
			if client.ChatID != "" {
				if _, ok := manager.chatRooms[client.ChatID]; !ok {
					manager.chatRooms[client.ChatID] = make(map[string]bool)
				}
				manager.chatRooms[client.ChatID][client.ID] = true
			}
			manager.mutex.Unlock()

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client.ID]; ok {
				delete(manager.clients, client.ID)
				delete(manager.userToClient, client.UserID)

				// Remove from chat room if chatID is provided
				if client.ChatID != "" {
					if _, ok := manager.chatRooms[client.ChatID]; ok {
						delete(manager.chatRooms[client.ChatID], client.ID)
						if len(manager.chatRooms[client.ChatID]) == 0 {
							delete(manager.chatRooms, client.ChatID)
						}
					}
				}

				close(client.Send)
			}
			manager.mutex.Unlock()

		case message := <-manager.broadcast:
			manager.mutex.RLock()
			if clients, ok := manager.chatRooms[message.ChatID]; ok {
				for clientID := range clients {
					if client, found := manager.clients[clientID]; found {
						select {
						case client.Send <- message.Message:
						default:
							manager.mutex.RUnlock()
							manager.mutex.Lock()
							delete(manager.clients, client.ID)
							if client.ChatID != "" {
								delete(manager.chatRooms[client.ChatID], client.ID)
								if len(manager.chatRooms[client.ChatID]) == 0 {
									delete(manager.chatRooms, client.ChatID)
								}
							}
							delete(manager.userToClient, client.UserID)
							close(client.Send)
							manager.mutex.Unlock()
							manager.mutex.RLock()
						}
					}
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// GetWebSocketManager returns the singleton instance of WebSocketManager
func GetWebSocketManager() *WebSocketManager {
	wsManagerOnce.Do(func() {
		wsManager = NewWebSocketManager()
		log.Println("WebSocket manager initialized")
	})
	return wsManager
}
