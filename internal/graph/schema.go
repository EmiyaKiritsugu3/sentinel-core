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
`

func Migrate(db *sqlite.DB) error {
	_, err := db.Conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not run migration: %w", err)
	}
	return nil
}
