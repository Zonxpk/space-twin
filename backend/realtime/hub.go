package realtime

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mutex     sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

func (h *Hub) Run() {
	// Simulate data loop
	go func() {
		for {
			time.Sleep(2 * time.Second)
			h.broadcastSimulation()
		}
	}()

	for {
		message := <-h.broadcast
		h.mutex.Lock()
		for client := range h.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Websocket error: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade:", err)
		return
	}

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	// Clean up on close
	defer func() {
		h.mutex.Lock()
		delete(h.clients, conn)
		h.mutex.Unlock()
		conn.Close()
	}()

	// Keep alive / read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Hub) broadcastSimulation() {
	// Simulate room data
	statuses := []string{"AVAILABLE", "BUSY", "OFFLINE"}

	updates := []map[string]interface{}{
		{"name": "Lobby", "status": statuses[rand.Intn(2)], "occupancy": rand.Intn(10)}, // AVAILABLE or BUSY
		{"name": "Office 101", "status": statuses[rand.Intn(3)], "occupancy": rand.Intn(2)},
		{"name": "Meeting Room A", "status": "BUSY", "occupancy": rand.Intn(5)}, // Usually busy
	}

	// Add random updates for other rooms potentially found
	// In real app, we would track rooms by ID.

	msg, _ := json.Marshal(map[string]interface{}{
		"type": "update",
		"data": updates,
	})

	h.broadcast <- msg
}
