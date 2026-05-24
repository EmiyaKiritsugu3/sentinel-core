package knowledge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

const debriefTemplate = `# Session Debrief — {{.Date}} {{.Time}}

## Decisions Made
{{range .Decisions}}- {{.Summary}}
{{end}}
## Patterns Observed
### Anti-Patterns (what failed)
{{range .Errors}}- {{.Summary}}
{{end}}
### Success Patterns (what worked)
{{range .Patterns}}- {{.Summary}}
{{end}}
## Files Changed
{{range .FileChanges}}- {{.File}} — {{.Summary}}
{{end}}
## Domain Tags
{{range .Domains}}- {{.}}
{{end}}
## Follow-ups
- [ ] ...
`

// DebriefData holds the structured template variables for debrief rendering.
type DebriefData struct {
	Date        string
	Time        string
	Decisions   []SessionEvent
	Errors      []SessionEvent
	Patterns    []SessionEvent
	FileChanges []SessionEvent
	Domains     []string
}

// DebriefService generates and persists session debriefs.
type DebriefService struct {
	buffer  *EventBuffer
	db      *sqlite.DB
	baseDir string
	tmpl    string
}

// NewDebriefService creates and returns a DebriefService configured with the provided
// EventBuffer, optional sqlite.DB, and base directory used to store generated session files.
// The baseDir is the knowledge root directory (typically ~/knowledge).
func NewDebriefService(buffer *EventBuffer, db *sqlite.DB, baseDir string) *DebriefService {
	return &DebriefService{
		buffer:  buffer,
		db:      db,
		baseDir: baseDir,
		tmpl:    debriefTemplate,
	}
}

// Generate renders the debrief markdown from the current buffer contents.
func (s *DebriefService) Generate() string {
	now := time.Now()
	domainSet := make(map[string]bool)
	for _, e := range s.buffer.Snapshot() {
		if e.Domain != "" {
			domainSet[e.Domain] = true
		}
	}
	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	data := DebriefData{
		Date:        now.Format("2006-01-02"),
		Time:        now.Format("15:04"),
		Decisions:   s.buffer.Decisions(),
		Errors:      s.buffer.Errors(),
		Patterns:    s.buffer.Patterns(),
		FileChanges: s.buffer.ByType(EventFileChange),
		Domains:     domains,
	}
	var buf bytes.Buffer
	tmpl, err := template.New("debrief").Parse(s.tmpl)
	if err != nil {
		return fmt.Sprintf("<!-- template error: %v -->\n%s", err, s.renderFallback(data))
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<!-- template error: %v -->\n%s", err, s.renderFallback(data))
	}
	return buf.String()
}

func (s *DebriefService) renderFallback(data DebriefData) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Session Debrief — %s %s\n\n", data.Date, data.Time))
	b.WriteString("## Decisions Made\n")
	for _, d := range data.Decisions {
		b.WriteString(fmt.Sprintf("- %s\n", d.Summary))
	}
	b.WriteString("\n## Patterns Observed\n### Anti-Patterns\n")
	for _, e := range data.Errors {
		b.WriteString(fmt.Sprintf("- %s\n", e.Summary))
	}
	b.WriteString("### Success Patterns\n")
	for _, p := range data.Patterns {
		b.WriteString(fmt.Sprintf("- %s\n", p.Summary))
	}
	b.WriteString("\n## Files Changed\n")
	for _, f := range data.FileChanges {
		b.WriteString(fmt.Sprintf("- %s — %s\n", f.File, f.Summary))
	}
	b.WriteString("\n## Domain Tags\n")
	for _, d := range data.Domains {
		b.WriteString(fmt.Sprintf("- %s\n", d))
	}
	b.WriteString("\n## Follow-ups\n- [ ] ...\n")
	return b.String()
}

// Save persists the debrief to the filesystem and graph database.
// Returns the session ID, markdown path, and any error.
func (s *DebriefService) Save(ctx context.Context) (string, string, error) {
	return s.SaveContent(ctx, s.Generate())
}

// SaveContent persists pre-rendered content to the filesystem and graph database.
// Unlike Save, it does not regenerate from the buffer — it uses the provided content as-is.
func (s *DebriefService) SaveContent(ctx context.Context, content string) (string, string, error) {
	now := time.Now()
	sessionID := uuid.New().String()[:8]
	filename := fmt.Sprintf("%s-%s.md", now.Format("2006-01-02"), now.Format("1504"))
	dir := filepath.Join(s.baseDir, "sessions")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("debrief: create sessions dir %s: %w", dir, err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", "", fmt.Errorf("debrief: write markdown: %w", err)
	}

	if s.db != nil {
		if err := s.saveToGraph(ctx, sessionID, path, now); err != nil {
			fmt.Fprintf(os.Stderr, "warning: debrief graph persistence failed: %v\n", err)
		}
	}

	return sessionID, path, nil
}

func (s *DebriefService) saveToGraph(ctx context.Context, sessionID, path string, now time.Time) error {
	if err := sqlite.ValidateDB(s.db, "debrief-graph"); err != nil {
		return err
	}
	tx, err := s.db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("debrief: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	decisions := s.buffer.Decisions()
	errors_ := s.buffer.Errors()
	patterns := s.buffer.Patterns()
	allEvents := s.buffer.Snapshot()
	domainSet := make(map[string]bool)
	for _, e := range allEvents {
		if e.Domain != "" {
			domainSet[e.Domain] = true
		}
	}
	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO knowledge_sessions (id, markdown_path, started_at, ended_at, event_count, decision_count, error_count, pattern_count, domains) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, path, now.Add(-time.Hour), now, len(allEvents), len(decisions), len(errors_), len(patterns), strings.Join(domains, ","),
	)
	if err != nil {
		return fmt.Errorf("debrief: insert session: %w", err)
	}
	for _, e := range allEvents {
		tags := strings.Join(e.Tags, ",")
		_, err = tx.ExecContext(ctx,
			`INSERT INTO session_events (session_id, event_type, domain, summary, detail, file_path, tags) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			sessionID, string(e.Type), e.Domain, e.Summary, e.Detail, e.File, tags,
		)
		if err != nil {
			return fmt.Errorf("debrief: insert event: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("debrief: commit tx: %w", err)
	}
	return nil
}
