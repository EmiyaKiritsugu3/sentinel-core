package patterns

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

const (
	CategoryAntiPattern = "anti-pattern"
	CategoryCognitivePattern = "cognitive-pattern"
	CategoryStructuralPrinciple = "structural-principle"
	CategoryRoutingPrinciple = "routing-principle"

	timeLayout = "2006-01-02 15:04:05"
)

const (
	SourceCognitiveDNA     = "cognitive-dna"
	SourceEvolutionInsights = "evolution-insights"
	SourceSentinelLog      = "sentinel-log"
	SourceManual           = "manual"
	SourceEpiphany         = "epiphany"
)

const (
	ImpactHigh   = "high"
	ImpactMedium = "medium"
	ImpactLow    = "low"
)

type Pattern struct {
	ID          string
	Title       string
	Description string
	Category    string
	Source      string
	SourceRef   string
	Tags        string
	Impact      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListFilters struct {
	Category string
	Source   string
	Impact   string
	Limit    int
}

type PatternStore struct {
	db *sqlite.DB
}

func NewPatternStore(db *sqlite.DB) (*PatternStore, error) {
	if err := sqlite.ValidateDB(db, "pattern-store"); err != nil {
		return nil, err
	}
	return &PatternStore{db: db}, nil
}

func (s *PatternStore) Create(p *Pattern) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO patterns (id, title, description, category, source, source_ref, tags, impact)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Conn.Exec(query, id, p.Title, p.Description, p.Category, p.Source, p.SourceRef, p.Tags, p.Impact)
	if err != nil {
		return "", fmt.Errorf("patterns: failed to create pattern: %w", err)
	}
	return id, nil
}

func (s *PatternStore) List(filters ListFilters) ([]Pattern, error) {
	query := "SELECT id, title, description, category, source, source_ref, tags, impact, created_at, updated_at FROM patterns WHERE 1=1"
	var args []interface{}

	if filters.Category != "" {
		query += " AND category = ?"
		args = append(args, filters.Category)
	}
	if filters.Source != "" {
		query += " AND source = ?"
		args = append(args, filters.Source)
	}
	if filters.Impact != "" {
		query += " AND impact = ?"
		args = append(args, filters.Impact)
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
	}

	rows, err := s.db.Conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("patterns: failed to list patterns: %w", err)
	}
	defer rows.Close()

	return scanPatterns(rows)
}

func (s *PatternStore) Search(query string) ([]Pattern, error) {
	q := `SELECT p.id, p.title, p.description, p.category, p.source, p.source_ref, p.tags, p.impact, p.created_at, p.updated_at
	FROM patterns p
	JOIN patterns_fts fts ON p.rowid = fts.rowid
	WHERE patterns_fts MATCH ?
	ORDER BY bm25(patterns_fts) DESC
	LIMIT 20`
	rows, err := s.db.Conn.Query(q, query)
	if err != nil {
		return nil, fmt.Errorf("patterns: search failed: %w", err)
	}
	defer rows.Close()

	return scanPatterns(rows)
}

func (s *PatternStore) Get(id string) (*Pattern, error) {
	query := `SELECT id, title, description, category, source, source_ref, tags, impact, created_at, updated_at
	FROM patterns WHERE id = ?`
	var p Pattern
	var createdAt, updatedAt string
	err := s.db.Conn.QueryRow(query, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.Category, &p.Source,
		&p.SourceRef, &p.Tags, &p.Impact, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("patterns: pattern %s not found: %w", id, err)
	}
	p.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	p.UpdatedAt, _ = time.Parse(timeLayout, updatedAt)
	return &p, nil
}

func scanPatterns(rows *sql.Rows) ([]Pattern, error) {
	var patterns []Pattern
	for rows.Next() {
		var p Pattern
		var createdAt, updatedAt string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Category, &p.Source,
			&p.SourceRef, &p.Tags, &p.Impact, &createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("patterns: failed to scan pattern: %w", err)
		}
		p.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		p.UpdatedAt, _ = time.Parse(timeLayout, updatedAt)
		patterns = append(patterns, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("patterns: row iteration error: %w", err)
	}
	return patterns, nil
}
