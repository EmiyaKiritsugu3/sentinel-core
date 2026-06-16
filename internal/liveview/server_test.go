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

func TestServer_ServeWS_OriginCheck(t *testing.T) {
	t.Parallel()
	server := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Run(ctx) }()

	ts := httptest.NewServer(http.HandlerFunc(server.serveWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	tests := []struct {
		name    string
		origin  string
		wantErr bool
	}{
		{
			name:    "valid localhost origin",
			origin:  "http://localhost:5173",
			wantErr: false,
		},
		{
			name:    "valid 127.0.0.1 origin",
			origin:  "http://127.0.0.1:8080",
			wantErr: false,
		},
		{
			name:    "no origin",
			origin:  "",
			wantErr: false, // Upgrader allows empty origin for non-browser clients
		},
		{
			name:    "invalid external origin",
			origin:  "https://example.com",
			wantErr: true,
		},
		{
			name:    "malicious local subdomain",
			origin:  "http://localhost.evil.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialer := websocket.Dialer{}
			var header http.Header
			if tt.origin != "" {
				header = http.Header{"Origin": []string{tt.origin}}
			}
			conn, resp, err := dialer.Dial(wsURL, header)

			if tt.wantErr {
				if err == nil {
					conn.Close()
					if resp != nil {
						resp.Body.Close()
					}
					t.Errorf("expected error for origin %q, got none", tt.origin)
				} else if resp != nil && resp.StatusCode != http.StatusForbidden {
					t.Errorf("expected status 403 for origin %q, got %d", tt.origin, resp.StatusCode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for origin %q: %v", tt.origin, err)
				}
				if conn != nil {
					conn.Close()
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}
}

func TestServer_StartHTTP_InvalidDB(t *testing.T) {
	t.Parallel()
	server := NewServer()

	// Create a DB but skip initialization, testing the ValidateDB error path
	err := server.StartHTTP(8080, nil)
	if err == nil {
		t.Errorf("expected error for nil DB")
	}
}
