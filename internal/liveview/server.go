package liveview

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := u.Hostname() // strips port, no prefix-match bypass
		return host == "localhost" || host == "127.0.0.1"
	},
}

// wsClient wraps a WebSocket connection with a dedicated send channel,
// ensuring only writePump ever writes to the connection.
type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

// Server acts as a WebSocket hub and implements graph.Observer
type Server struct {
	clients    map[*wsClient]bool
	broadcast  chan graph.GraphEvent
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		broadcast:  make(chan graph.GraphEvent, 256), // Buffer to prevent engine blocking
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		clients:    make(map[*wsClient]bool),
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
		case c := <-s.register:
			s.mu.Lock()
			s.clients[c] = true
			s.mu.Unlock()
		case c := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[c]; ok {
				delete(s.clients, c)
				close(c.send)
				c.conn.Close()
			}
			s.mu.Unlock()
		case event := <-s.broadcast:
			payload, err := json.Marshal(event)
			if err != nil {
				log.Printf("liveview: failed to serialize event %v: %v\n", event.Type, err)
				continue
			}

			s.mu.RLock()
			for c := range s.clients {
				select {
				case c.send <- payload:
				default:
					log.Printf("liveview: client send buffer full, dropping event\n")
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

	c := &wsClient{conn: conn, send: make(chan []byte, 256)}
	s.register <- c

	go s.readPump(c)
	go s.writePump(c)
}

func (s *Server) readPump(c *wsClient) {
	defer func() {
		s.unregister <- c
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("liveview: read error: %v", err)
			}
			break
		}
	}
}

// writePump is the sole writer goroutine for a client connection.
// Gorilla WebSocket requires that no more than one goroutine calls write methods concurrently.
func (s *Server) writePump(c *wsClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

	return http.ListenAndServe(addr, nil) // nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- local-only dev tool, TLS not applicable
}
