package graph

import (
	"fmt"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS nodes (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL, -- file, function, struct, interface
	file_path TEXT NOT NULL,
	start_line INTEGER,
	end_line INTEGER,
	hash TEXT,
	last_indexed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edges (
	from_node_id TEXT NOT NULL,
	to_node_id TEXT NOT NULL,
	relation_type TEXT NOT NULL, -- imports, calls, implements
	PRIMARY KEY (from_node_id, to_node_id, relation_type),
	FOREIGN KEY (from_node_id) REFERENCES nodes(id) ON DELETE CASCADE,
	FOREIGN KEY (to_node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	description TEXT NOT NULL,
	status TEXT NOT NULL, -- PENDING, IN_PROGRESS, AUDITING, DONE, FAILED
	tier TEXT, -- T1, T2, T3
	verification_command TEXT, -- O comando que o Audit Runner deve executar
	commit_hash TEXT,
	latency_ms REAL DEFAULT 0,
	tokens_used INTEGER DEFAULT 0,
	api_cost REAL DEFAULT 0,
	math_delta REAL DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT NOT NULL,
	command TEXT NOT NULL,
	output TEXT,
	exit_code INTEGER,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS standards (
        id TEXT PRIMARY KEY,
        description TEXT NOT NULL,
        pattern_query TEXT, -- Regra semântica ou regex
        status TEXT NOT NULL DEFAULT 'PROPOSED', -- PROPOSED, AUDITED, SEALED
        confidence_score INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS specialist_registry (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    base_persona TEXT NOT NULL,
    current_persona_path TEXT NOT NULL,
    parent_specialist_id TEXT,
    generation INTEGER DEFAULT 1,
    reliability_score REAL DEFAULT 1.0,
    success_rate REAL DEFAULT 0.0,
    tasks_completed INTEGER DEFAULT 0,
    capabilities TEXT, -- JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_specialist_id) REFERENCES specialist_registry(id)
);

CREATE TABLE IF NOT EXISTS sub_tasks (
    id TEXT PRIMARY KEY,
    parent_task_id TEXT NOT NULL,
    specialist_id TEXT,
    description TEXT NOT NULL,
    status TEXT NOT NULL, -- PENDING, DISPATCHED, IN_PROGRESS, AUDITING, DONE, FAILED
    worktree_path TEXT,
    branch_name TEXT,
    heartbeat TIMESTAMP,
    required_capabilities TEXT, -- JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (specialist_id) REFERENCES specialist_registry(id)
);

CREATE TABLE IF NOT EXISTS performance_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    specialist_id TEXT,
    task_id TEXT,
    sub_task_id TEXT,
    token_usage INTEGER,
    duration_ms INTEGER,
    audit_passed BOOLEAN,
    failure_reason TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (specialist_id) REFERENCES specialist_registry(id),
    FOREIGN KEY (sub_task_id) REFERENCES sub_tasks(id)
);
`

func Migrate(db *sqlite.DB) error {
	_, err := db.Conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not run migration: %w", err)
	}

	// Migrations for SME Phase 1 metrics
	migrations := []string{
		"ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;",
		"ALTER TABLE tasks ADD COLUMN tokens_used INTEGER DEFAULT 0;",
		"ALTER TABLE tasks ADD COLUMN api_cost REAL DEFAULT 0;",
		"ALTER TABLE tasks ADD COLUMN math_delta REAL DEFAULT 0;",
	}
	for _, m := range migrations {
		// Ignore errors since columns might already exist
		_, _ = db.Conn.Exec(m)
	}

	// KISS Specialist Seeding
	seeds := []struct {
		id   string
		name string
		caps string
	}{
		{"spec-go", "Go Specialist", `["go", "test", "ast"]`},
		{"spec-md", "Documentation Specialist", `["markdown", "adr"]`},
		{"spec-git", "VCS Specialist", `["git", "worktree"]`},
	}

	for _, s := range seeds {
		query := "INSERT OR IGNORE INTO specialist_registry (id, name, base_persona, current_persona_path, capabilities) VALUES (?, ?, ?, ?, ?)"
		if _, err := db.Conn.Exec(query, s.id, s.name, "Base", "internal/agents/definitions/architect.md", s.caps); err != nil {
			return fmt.Errorf("migrate: failed to seed specialist %s: %w", s.id, err)
		}
	}

	return nil
}
