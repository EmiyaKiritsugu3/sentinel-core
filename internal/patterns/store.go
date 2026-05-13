package patterns

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

// CategoryAntiPattern identifies the anti-pattern category.
const CategoryAntiPattern = "anti-pattern"

// CategoryCognitivePattern identifies the cognitive-pattern category.
const CategoryCognitivePattern = "cognitive-pattern"

// CategoryStructuralPrinciple identifies the structural-principle category.
const CategoryStructuralPrinciple = "structural-principle"

// CategoryRoutingPrinciple identifies the routing-principle category.
const CategoryRoutingPrinciple = "routing-principle"

const timeLayout = "2006-01-02 15:04:05"

// SourceCognitiveDNA identifies cognitive DNA as the pattern source.
const SourceCognitiveDNA = "cognitive-dna"

// SourceEvolutionInsights identifies evolution insights as the pattern source.
const SourceEvolutionInsights = "evolution-insights"

// SourceSentinelLog identifies the sentinel log as the pattern source.
const SourceSentinelLog = "sentinel-log"

// SourceManual identifies manual input as the pattern source.
const SourceManual = "manual"

// SourceEpiphany identifies epiphany as the pattern source.
const SourceEpiphany = "epiphany"

// ImpactHigh represents a high impact level.
const ImpactHigh = "high"

// ImpactMedium represents a medium impact level.
const ImpactMedium = "medium"

// ImpactLow represents a low impact level.
const ImpactLow = "low"

// Pattern represents an architectural pattern with metadata.
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

// ListFilters holds filter parameters for listing patterns.
type ListFilters struct {
	Category string
	Source   string
	Impact   string
	Limit    int
}

// PatternStore persists and retrieves patterns from a SQLite database.
type PatternStore struct {
	db *sqlite.DB
}

// NewPatternStore creates a new PatternStore with the given database connection.
func NewPatternStore(db *sqlite.DB) (*PatternStore, error) {
	if err := sqlite.ValidateDB(db, "pattern-store"); err != nil {
		return nil, err
	}
	return &PatternStore{db: db}, nil
}

// Create inserts a new pattern and returns its generated ID.
func (s *PatternStore) Create(ctx context.Context, p *Pattern) (string, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.Create"); err != nil {
		return "", err
	}
	if p == nil {
		return "", fmt.Errorf("pattern-store.Create: nil pattern")
	}
	id := uuid.New().String()
	query := `INSERT INTO patterns (id, title, description, category, source, source_ref, tags, impact)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Conn.ExecContext(ctx, query, id, p.Title, p.Description, p.Category, p.Source, p.SourceRef, p.Tags, p.Impact)
	if err != nil {
		return "", fmt.Errorf("patterns: failed to create pattern: %w", err)
	}
	return id, nil
}

// List returns patterns matching the given filters, ordered by creation date descending.
func (s *PatternStore) List(ctx context.Context, filters ListFilters) ([]Pattern, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.List"); err != nil {
		return nil, err
	}
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
		query += fmt.Sprintf(" LIMIT %d", filters.Limit) //nolint:gosec // filters.Limit is validated integer
	}

	rows, err := s.db.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("patterns: failed to list patterns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPatterns(rows)
}

// Search performs a full-text search on patterns using the given query.
func (s *PatternStore) Search(ctx context.Context, query string) ([]Pattern, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.Search"); err != nil {
		return nil, err
	}
	q := `SELECT p.id, p.title, p.description, p.category, p.source, p.source_ref, p.tags, p.impact, p.created_at, p.updated_at
	FROM patterns p
	JOIN patterns_fts fts ON p.rowid = fts.rowid
	WHERE patterns_fts MATCH ?
	ORDER BY bm25(patterns_fts) ASC
	LIMIT 20`
	rows, err := s.db.Conn.QueryContext(ctx, q, query)
	if err != nil {
		return nil, fmt.Errorf("patterns: search failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPatterns(rows)
}

// Get retrieves a single pattern by its ID.
func (s *PatternStore) Get(ctx context.Context, id string) (*Pattern, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.Get"); err != nil {
		return nil, err
	}
	query := `SELECT id, title, description, category, source, source_ref, tags, impact, created_at, updated_at
	FROM patterns WHERE id = ?`
	var p Pattern
	var createdAt, updatedAt string
	err := s.db.Conn.QueryRowContext(ctx, query, id).Scan(
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
