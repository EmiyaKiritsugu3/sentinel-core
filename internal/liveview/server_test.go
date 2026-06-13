package liveview

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/gorilla/websocket"
)

func TestServer_Broadcast(t *testing.T) {
	t.Parallel()
	server := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Run(ctx) }()

	// Start test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(server.serveWS))
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	defer func() { _ = conn.Close() }()

	// Poll until the hub has registered the client instead of sleeping a fixed duration.
	deadline := time.After(2 * time.Second)
	for {
		server.mu.RLock()
		registered := len(server.clients) > 0
		server.mu.RUnlock()
		if registered {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for client registration")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	// Send an event
	event := graph.GraphEvent{
		Type: graph.EventScanStarted,
		Time: time.Now(),
	}
	server.Notify(event)

	// Read event from client
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var receivedEvent graph.GraphEvent
	err = json.Unmarshal(msg, &receivedEvent)
	if err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if receivedEvent.Type != graph.EventScanStarted {
		t.Errorf("expected event type %v, got %v", graph.EventScanStarted, receivedEvent.Type)
	}
}

func TestServer_ConcurrentNotify(t *testing.T) {
	t.Parallel()
	server := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Run(ctx) }()

	var wg sync.WaitGroup

	// Simulate many concurrent engine notifications
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Notify(graph.GraphEvent{
				Type: graph.EventNodeUpserted,
			})
		}()
	}

	wg.Wait()
	// Should not block or panic
}

func TestUpgraderCheckOrigin(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		want       bool
	}{
		{
			name:       "No Origin",
			origin:     "",
			want:       true,
		},
		{
			name:       "Invalid URL",
			origin:     "://bad-url",
			want:       false,
		},
		{
			name:       "Allowed Localhost",
			origin:     "http://localhost:5173",
			want:       true,
		},
		{
			name:       "Allowed 127.0.0.1",
			origin:     "http://127.0.0.1:3000",
			want:       true,
		},
		{
			name:       "Disallowed Origin",
			origin:     "http://evil.com",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			got := upgrader.CheckOrigin(req)
			if got != tt.want {
				t.Errorf("expected CheckOrigin %v, got %v", tt.want, got)
			}
		})
	}
}
