package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"
)

type Event struct {
	Type   string      `json:"type"`
	PageID uint        `json:"page_id,omitempty"`
	UserID uint        `json:"user_id,omitempty"`
	Data   interface{} `json:"data"`
}

type Hub struct {
	mu               sync.RWMutex
	connections      map[uint]map[*gws.Conn]bool // pageID -> set of connections
	userConnections  map[uint]map[*gws.Conn]bool // userID -> set of connections
	upgrader         gws.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		connections:     make(map[uint]map[*gws.Conn]bool),
		userConnections: make(map[uint]map[*gws.Conn]bool),
		upgrader: gws.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Check for user_id first (user-level connection)
	userIDStr := r.URL.Query().Get("user_id")
	userID64, _ := strconv.ParseUint(userIDStr, 10, 32)
	userID := uint(userID64)

	if userID != 0 {
		h.registerUser(userID, conn)
		defer h.unregisterUser(userID, conn)
	} else {
		// Fall back to page-level connection
		pageIDStr := r.URL.Query().Get("page_id")
		pageID64, _ := strconv.ParseUint(pageIDStr, 10, 32)
		pageID := uint(pageID64)

		h.register(pageID, conn)
		defer h.unregister(pageID, conn)
	}

	// Set up ping/pong handler to keep connection alive
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker in goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			if err := conn.WriteMessage(gws.PingMessage, nil); err != nil {
				return
			}
			<-ticker.C
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Hub) register(pageID uint, conn *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.connections[pageID]; !ok {
		h.connections[pageID] = make(map[*gws.Conn]bool)
	}
	h.connections[pageID][conn] = true
}

func (h *Hub) unregister(pageID uint, conn *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.connections[pageID]; ok {
		delete(h.connections[pageID], conn)
		if len(h.connections[pageID]) == 0 {
			delete(h.connections, pageID)
		}
	}
}

func (h *Hub) registerUser(userID uint, conn *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.userConnections[userID]; !ok {
		h.userConnections[userID] = make(map[*gws.Conn]bool)
	}
	h.userConnections[userID][conn] = true
	log.Printf("[ws] user %d connected", userID)
}

func (h *Hub) unregisterUser(userID uint, conn *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.userConnections[userID]; ok {
		delete(h.userConnections[userID], conn)
		if len(h.userConnections[userID]) == 0 {
			delete(h.userConnections, userID)
		}
	}
	log.Printf("[ws] user %d disconnected", userID)
}

func (h *Hub) Broadcast(pageID uint, evt Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload, _ := json.Marshal(evt)
	log.Printf("[ws] Broadcasting event type=%s pageID=%d to %d page connections", evt.Type, pageID, len(h.connections[pageID]))
	
	// Broadcast to specific page connections
	for conn := range h.connections[pageID] {
		if err := conn.WriteMessage(gws.TextMessage, payload); err != nil {
			log.Printf("[ws] Failed to send to page connection: %v", err)
		}
	}

	// Broadcast to global connections (pageID = 0)
	if pageID != 0 {
		for conn := range h.connections[0] {
			if err := conn.WriteMessage(gws.TextMessage, payload); err != nil {
				log.Printf("[ws] Failed to send to global connection: %v", err)
			}
		}
	}

	// Broadcast to ALL user connections (for inbox updates)
	log.Printf("[ws] Broadcasting to %d user connections", len(h.userConnections))
	for userID, conns := range h.userConnections {
		for conn := range conns {
			if err := conn.WriteMessage(gws.TextMessage, payload); err != nil {
				log.Printf("[ws] Failed to send to user %d: %v", userID, err)
			}
		}
	}
}

// BroadcastUser sends an event to all connections for a specific user.
func (h *Hub) BroadcastUser(userID uint, evt interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[ws] failed to marshal user event: %v", err)
		return
	}

	for conn := range h.userConnections[userID] {
		if err := conn.WriteMessage(gws.TextMessage, payload); err != nil {
			log.Printf("[ws] failed to send to user %d: %v", userID, err)
		}
	}
}

// BroadcastBadgeUpdate sends a badge count update to a user.
func (h *Hub) BroadcastBadgeUpdate(userID uint, count int64) {
	h.BroadcastUser(userID, map[string]interface{}{
		"type":   "BADGE_UPDATE",
		"user_id": userID,
		"count":  count,
	})
}
