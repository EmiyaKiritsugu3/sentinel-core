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

func TestServer_StartHTTP_InvalidDB(t *testing.T) {
	s := NewServer()
	err := s.StartHTTP(8080, nil)
	if err == nil || !strings.Contains(err.Error(), "liveview: nil db") {
		t.Fatalf("expected error for nil db, got %v", err)
	}
}

func TestServer_ServeWS_NoUpgrade(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	s.serveWS(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestServer_RunContext(t *testing.T) {
	s := NewServer()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		s.Notify(graph.GraphEvent{Type: "TEST"})
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := s.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestServer_ServeWS_ValidOrigin(t *testing.T) {
	checkOrigin := upgrader.CheckOrigin
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	if !checkOrigin(req) {
		t.Fatalf("expected origin to be allowed")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req2.Header.Set("Origin", "http://malicious.com")

	if checkOrigin(req2) {
		t.Fatalf("expected origin to be rejected")
	}

	req3 := httptest.NewRequest(http.MethodGet, "/ws", nil)

	if !checkOrigin(req3) {
		t.Fatalf("expected empty origin to be allowed")
	}

	req4 := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req4.Header.Set("Origin", "::1%invalid")

	if checkOrigin(req4) {
		t.Fatalf("expected invalid origin to be rejected")
	}
}
