package graph

import (
	"database/sql"
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

CREATE TABLE IF NOT EXISTS agent_trust (
    agent_name  TEXT PRIMARY KEY,
    successes   INTEGER NOT NULL DEFAULT 0,
    total       INTEGER NOT NULL DEFAULT 0,
    trust_score REAL NOT NULL DEFAULT 0.5,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

func Migrate(db *sqlite.DB) (err error) {
	if err := sqlite.ValidateDB(db, "migrate"); err != nil {
		return err
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("migrate: could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not run migration schema: %w", err)
	}

	// Migrations for SME Phase 1 metrics — guarded by column-existence checks
	// to avoid swallowing unrelated SQLite errors via substring matching.
	type colMig struct {
		table, column, sql string
	}
	for _, m := range []colMig{
		{"tasks", "latency_ms", "ALTER TABLE tasks ADD COLUMN latency_ms REAL DEFAULT 0;"},
		{"tasks", "tokens_used", "ALTER TABLE tasks ADD COLUMN tokens_used INTEGER DEFAULT 0;"},
		{"tasks", "api_cost", "ALTER TABLE tasks ADD COLUMN api_cost REAL DEFAULT 0;"},
		{"tasks", "math_delta", "ALTER TABLE tasks ADD COLUMN math_delta REAL DEFAULT 0;"},
	} {
	exists, err := columnExistsInTx(tx, m.table, m.column)
	if err != nil {
		return fmt.Errorf("migrate: checking %s.%s: %w", m.table, m.column, err)
	}
		if !exists {
			if _, err = tx.Exec(m.sql); err != nil {
				return fmt.Errorf("migrate: %s: %w", m.sql, err)
			}
		}
	}

	// Rename specialist_id → agent_name if the old column name still exists.
oldExists, err := columnExistsInTx(tx, "agent_trust", "specialist_id")
if err != nil {
	return fmt.Errorf("migrate: checking agent_trust.specialist_id: %w", err)
}
	if oldExists {
		if _, err = tx.Exec("ALTER TABLE agent_trust RENAME COLUMN specialist_id TO agent_name;"); err != nil {
			return fmt.Errorf("migrate: rename specialist_id: %w", err)
		}
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
		if _, err = tx.Exec(query, s.id, s.name, "Base", "internal/agents/definitions/architect.md", s.caps); err != nil {
			return fmt.Errorf("migrate: failed to seed specialist %s: %w", s.id, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("migrate: failed to commit transaction: %w", err)
	}

	return nil
}

// pragmaTableInfo maps known schema tables to their PRAGMA queries,
// avoiding dynamic SQL construction (SonarCloud rule S2077).
var pragmaTableInfo = map[string]string{
	"tasks":               "PRAGMA table_info(tasks)",
	"agent_trust":         "PRAGMA table_info(agent_trust)",
	"specialist_registry": "PRAGMA table_info(specialist_registry)",
	"sub_tasks":           "PRAGMA table_info(sub_tasks)",
	"nodes":               "PRAGMA table_info(nodes)",
	"edges":               "PRAGMA table_info(edges)",
	"audit_logs":          "PRAGMA table_info(audit_logs)",
	"standards":           "PRAGMA table_info(standards)",
	"performance_logs":    "PRAGMA table_info(performance_logs)",
}

func columnExistsInTx(tx *sql.Tx, table, column string) (bool, error) {
	query, ok := pragmaTableInfo[table]
	if !ok {
		return false, fmt.Errorf("migrate: unknown table %q", table)
	}
	rows, err := tx.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
