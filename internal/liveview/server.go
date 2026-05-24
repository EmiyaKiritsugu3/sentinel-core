package liveview

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

// NewServer creates a new WebSocket hub Server.
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
				_ = c.conn.Close()
			}
			s.mu.Unlock()
		case event := <-s.broadcast:
			payload, err := json.Marshal(event)
			if err != nil {
				slog.Warn("failed to serialize event", "type", event.Type, "error", err)
				continue
			}

			s.mu.RLock()
			for c := range s.clients {
				select {
				case c.send <- payload:
				default:
					slog.Warn("client send buffer full, dropping event")
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
		slog.Warn("broadcast channel full, dropping event", "type", event.Type)
	}
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket upgrade error", "error", err)
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
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("read error", "error", err)
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
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// StartHTTP starts the blocking HTTP server
func (s *Server) StartHTTP(port int, db *sqlite.DB) error {
	if err := sqlite.ValidateDB(db, "liveview"); err != nil {
		return fmt.Errorf("liveview: %w", err)
	}

	http.HandleFunc("/ws", s.serveWS)
	http.HandleFunc("/api/graph", handleGetGraph(db))
	http.HandleFunc("/api/status", handleGetStatus(db))
	http.HandleFunc("/api/code", handleGetCode(db))
	http.HandleFunc("/api/adr", handleListADR(db))
	http.HandleFunc("/api/adr/", handleGetADR(db))

	// Serve the Vite build
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	slog.Info("liveview server listening", "addr", addr)

	return http.ListenAndServe(addr, nil) //nolint:gosec // nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- local-only dev tool, TLS not applicable
}
