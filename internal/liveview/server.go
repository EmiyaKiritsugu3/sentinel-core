package liveview

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for the local dev tool
	},
}

// Server acts as a WebSocket hub and implements graph.Observer
type Server struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan graph.GraphEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		broadcast:  make(chan graph.GraphEvent, 256), // Buffer to prevent engine blocking
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

// Run starts the internal hub logic for managing connections
func (s *Server) Run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("liveview: hub panic: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Close()
			}
			s.mu.Unlock()
		case event := <-s.broadcast:
			// Serializa o evento aqui para não repetir por cliente (Performance)
			payload, err := json.Marshal(event)
			if err != nil {
				// Standard #05: Explicit error handling, do not panic
				log.Printf("liveview: failed to serialize event %v: %v\n", event.Type, err)
				continue
			}

			s.mu.RLock()
			for client := range s.clients {
				client.SetWriteDeadline(time.Now().Add(writeWait))
				err := client.WriteMessage(websocket.TextMessage, payload)
				if err != nil {
					log.Printf("liveview: write failed: %v\n", err)
					client.Close()
					// DO NOT delete from map while iterating with RLock.
					// The readPump will fail and send to unregister channel.
				}
			}
			s.mu.RUnlock()
		}
	}
}

// Notify is called by the graph.Engine when a new event occurs.
// It pushes the event to the broadcast channel non-blockingly.
func (s *Server) Notify(event graph.GraphEvent) {
	select {
	case s.broadcast <- event:
	default:
		log.Printf("liveview: broadcast channel full, dropping event %v\n", event.Type)
	}
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("liveview: upgrade error:", err)
		return
	}

	s.register <- conn

	// Inicia as rotinas de leitura/escrita para este cliente (ping/pong)
	go s.readPump(conn)
	go s.writePump(conn)
}

func (s *Server) readPump(client *websocket.Conn) {
	defer func() {
		s.unregister <- client
	}()

	client.SetReadLimit(512)
	client.SetReadDeadline(time.Now().Add(pongWait))
	client.SetPongHandler(func(string) error { client.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := client.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("liveview: read error: %v", err)
			}
			break
		}
	}
}

func (s *Server) writePump(client *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Close()
	}()

	for {
		select {
		case <-ticker.C:
			client.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// StartHTTP inicia o servidor HTTP bloqueante
func (s *Server) StartHTTP(port int, db *sqlite.DB) error {
	http.HandleFunc("/ws", s.serveWS)
	http.HandleFunc("/api/graph", handleGetGraph(db))

	// Servir o build do Vite
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("📡 Live View Server listening on %s...\n", addr)

	return http.ListenAndServe(addr, nil)
}
