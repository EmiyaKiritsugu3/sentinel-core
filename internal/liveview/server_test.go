package liveview

import (
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
	server := NewServer()
	go server.Run()

	// Start test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(server.serveWS))
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Give time for registration to process
	time.Sleep(50 * time.Millisecond)

	// Send an event
	event := graph.GraphEvent{
		Type: graph.EventScanStarted,
		Time: time.Now(),
	}
	server.Notify(event)

	// Read event from client
	conn.SetReadDeadline(time.Now().Add(time.Second))
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
	server := NewServer()
	go server.Run()

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
