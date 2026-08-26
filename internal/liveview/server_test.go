package liveview

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
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

func TestServer_StartHTTP_ValidDB(t *testing.T) {
	t.Parallel()
	server := NewServer()

	// Start a real server but tell it to close immediately
	// Note: We need a valid DB for this. Let's create an in-memory one with valid schema.
	// Wait, StartHTTP blocks. So we must run it in a goroutine and then shut it down.
	// But since this is a unit test that we just want coverage for,
	// we will mock it slightly. The simplest way to test `ListenAndServe`
	// is to give it an invalid port so it errors out immediately instead of blocking.

	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = rawDB.Close() }()
	db := &sqlite.DB{Conn: rawDB}
	// the validation requires "nodes", "edges", "tasks" tables
	createTasksTable(t, rawDB)
	_, _ = rawDB.Exec(`CREATE TABLE nodes(id TEXT PRIMARY KEY, name TEXT, type TEXT, file_path TEXT, start_line INTEGER, end_line INTEGER, hash TEXT, last_indexed TIMESTAMP)`)
	_, _ = rawDB.Exec(`CREATE TABLE edges(from_node_id TEXT, to_node_id TEXT, relation_type TEXT)`)

	// Port -1 should fail immediately on listen
	err = server.StartHTTP(-1, db)
	if err == nil {
		t.Fatal("expected error with invalid port")
	}
}

func TestServer_StartHTTP_InvalidDB(t *testing.T) {
	t.Parallel()
	server := NewServer()

	// db param is nil
	err := server.StartHTTP(0, nil)
	if err == nil {
		t.Fatal("expected error with nil db")
	}
	if !strings.Contains(err.Error(), "liveview: nil db") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
