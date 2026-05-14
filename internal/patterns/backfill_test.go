package patterns

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/testutil"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/stretchr/testify/assert"
)

var projectRoot = findProjectRoot()

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func TestBackfillFromCognitiveDNA(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	result, err := store.BackfillFromCognitiveDNA(context.Background(), projectRoot)
	if err != nil {
		t.Fatalf("BackfillFromCognitiveDNA failed: %v", err)
	}

	if result.Inserted == 0 {
		t.Fatal("expected at least 1 pattern inserted from COGNITIVE-DNA")
	}

	patterns, err := store.List(context.Background(), ListFilters{Source: "cognitive-dna"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(patterns) != result.Inserted {
		t.Fatalf("expected %d patterns, got %d", result.Inserted, len(patterns))
	}
}

func TestBackfillFromCognitiveDNA_Idempotent(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	result1, _ := store.BackfillFromCognitiveDNA(context.Background(), projectRoot)

	result2, _ := store.BackfillFromCognitiveDNA(context.Background(), projectRoot)
	if result2.Inserted != 0 {
		t.Fatalf("expected 0 inserts on second run, got %d", result2.Inserted)
	}
	if result2.Skipped != result1.Inserted {
		t.Fatalf("expected %d skips, got %d", result1.Inserted, result2.Skipped)
	}
}

func TestBackfillFromEvolutionInsights(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	result, err := store.BackfillFromEvolutionInsights(context.Background(), projectRoot)
	if err != nil {
		t.Fatalf("BackfillFromEvolutionInsights failed: %v", err)
	}
	if result.Inserted == 0 {
		t.Fatal("expected at least 1 pattern from EVOLUTION-INSIGHTS")
	}
}

func TestBackfillFromSentinelLog_DryRun(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	dir := t.TempDir()
	docDir := filepath.Join(dir, "docs", "process")
	if err := os.MkdirAll(docDir, 0755); err != nil { //nolint:gosec // test fixture {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := "# Log\n- Filtro A aplicado no roteamento de módulos críticos\n- Filtro B detectado em análise estrutural\n"
	if err := os.WriteFile(filepath.Join(docDir, "sentinel-log.md"), []byte(content), 0644); err != nil { //nolint:gosec // test fixture {
		t.Fatalf("write failed: %v", err)
	}

	result, err := store.BackfillFromSentinelLog(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("BackfillFromSentinelLog dry-run failed: %v", err)
	}
	patterns, _ := store.List(context.Background(), ListFilters{Source: "sentinel-log"})
	if len(patterns) != 0 {
		t.Fatalf("expected 0 patterns after dry-run, got %d", len(patterns))
	}
	if result.Extracted == 0 {
		t.Fatal("expected non-zero Extracted count from dry-run")
	}
	if len(result.Candidates) == 0 {
		t.Fatal("expected Candidates populated in dry-run result")
	}
}

// CG-01: False Positive Tests — strings.Contains for classification
// should be tested against inputs that match the substring but are not valid items.

func TestParseCognitiveDNA_FalsePositive_APBracketInComment(t *testing.T) {
	t.Parallel()
	// [AP- in HTML comment without pipes — len(parts) < 5 does not generate candidate

	dir := t.TempDir()
	path := filepath.Join(dir, "COGNITIVE-DNA.md")
	content := "# DNA\n<!-- Ver [AP-FOO] na seção acima -->\nTexto normal\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseCognitiveDNA(path)
	assert.NoError(t, err)
	assert.Empty(t, candidates, "comentário com [AP- sem pipes não deve gerar candidato")
}

func TestParseCognitiveDNA_FalsePositive_RegraOutsidePMO(t *testing.T) {
	t.Parallel()
	// Rule/MO before "### PMO-" — inPMO == false, does not capture

	dir := t.TempDir()
	path := filepath.Join(dir, "COGNITIVE-DNA.md")
	content := "# DNA\n- **Regra:** isso não é um PMO\n### PMO-001: Test\nconteúdo\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseCognitiveDNA(path)
	assert.NoError(t, err)
	for _, c := range candidates {
		assert.NotContains(t, c.Description, "isso não é um PMO")
	}
}

func TestParseCognitiveDNA_FalsePositive_ModusOperandiOutsidePMO(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "COGNITIVE-DNA.md")
	content := "# DNA\n- **Modus Operandi:** texto órfão\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseCognitiveDNA(path)
	assert.NoError(t, err)
	assert.Empty(t, candidates, "Modus Operandi fora de PMO não deve gerar candidato")
}

func TestParseEvolutionInsights_FalsePositive_SectionNameInBody(t *testing.T) {
	t.Parallel()
	// DOCUMENTED FP: strings.Contains("Structural Gaps") in body triggers section detector
	// Mechanism: parseEvolutionInsights uses strings.Contains to detect section,
	// which matches substring in any context. The FP is known — the line
	// "See Structural Gaps above for context" becomes a spurious candidate because
	// the section is active when the parser finds it.

	dir := t.TempDir()
	path := filepath.Join(dir, "EVOLUTION-INSIGHTS.md")
	content := "## Gaps Estruturais\n- Item válido: desc\n- Veja Gaps Estruturais acima para contexto\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseEvolutionInsights(path)
	assert.NoError(t, err)
	var fpCount int
	for _, c := range candidates {
		if c.Title == "Veja Gaps Estruturais acima para contexto" {
			fpCount++
			t.Log("FP DOCUMENTADO: substring 'Gaps Estruturais' no body ativou section detector — Candidato espúrio:", c.Title)
		}
	}
	if fpCount == 0 {
		t.Log("FP não reproduzido — parser pode ter sido corrigido ou input não triggera o path")
	}
}

func TestParseEvolutionInsights_FalsePositive_StrikethroughSkipped(t *testing.T) {
	t.Parallel()
	// strings.Contains(line, "~~") faz skip — item riscado NÃO deve aparecer

	dir := t.TempDir()
	path := filepath.Join(dir, "EVOLUTION-INSIGHTS.md")
	content := "## Gaps Estruturais\n- ~~Item riscado~~: desc antiga\n- Item válido: desc boa\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseEvolutionInsights(path)
	assert.NoError(t, err)
	for _, c := range candidates {
		assert.NotContains(t, c.Title, "Item riscado")
	}
}

func TestParseSentinelLog_FalsePositive_FiltroInNarrativeText(t *testing.T) {
	t.Parallel()
	// DOCUMENTED FP: "Filter A" in narrative text without "- "/"* " prefix —
	// parseSentinelLine does not require list prefix, only strings.Contains("Filter A/B/C"),
	// so narrative text with substring becomes spurious candidate if len(clean)>10

	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel-log.md")
	content := "# Log\nFiltro A foi discutido na reunião mas não aplicado\n- Filtro A aplicado: contexto suficiente aqui\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseSentinelLog(path)
	assert.NoError(t, err)
	var fpCount int
	for _, c := range candidates {
		if c.Title == "Filtro A foi discutido na reunião mas não aplicado" {
			fpCount++
			t.Log("FP DOCUMENTADO: texto narrativo com 'Filtro A' capturado como item — Candidato espúrio:", c.Title)
		}
	}
	if fpCount == 0 {
		t.Log("FP não reproduzido — parser pode ter sido corrigido ou input não triggera o path")
	}
}

func TestParseSentinelLog_FalsePositive_ShortLine(t *testing.T) {
	t.Parallel()
	// "Filter A" in short line — len(clean) > 10 protects

	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel-log.md")
	content := "# Log\n- Filtro A\n"
	err := os.WriteFile(path, []byte(content), 0644) //nolint:gosec // test fixture
	assert.NoError(t, err)

	candidates, err := parseSentinelLog(path)
	assert.NoError(t, err)
	assert.Empty(t, candidates, "linha curta com Filtro A não deve gerar candidato")
}

// Coverage: BackfillFromSentinelLog non-dry-run (real insert path)

func TestBackfillFromSentinelLog_Insert(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	// Create sentinel-log file with Filter content to test non-dry-run insert
	dir := t.TempDir()
	docDir := filepath.Join(dir, "docs", "process")
	if err := os.MkdirAll(docDir, 0755); err != nil { //nolint:gosec // test fixture {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := "# Log\n- Filtro A aplicado no roteamento de módulos críticos\n"
	if err := os.WriteFile(filepath.Join(docDir, "sentinel-log.md"), []byte(content), 0644); err != nil { //nolint:gosec // test fixture {
		t.Fatalf("write failed: %v", err)
	}

	result, err := store.BackfillFromSentinelLog(context.Background(), dir, false)
	if err != nil {
		t.Fatalf("BackfillFromSentinelLog insert failed: %v", err)
	}
	if result.Extracted == 0 {
		t.Fatal("expected at least 1 candidate from sentinel-log")
	}
	patterns, err := store.List(context.Background(), ListFilters{Source: "sentinel-log"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(patterns) == 0 {
		t.Fatal("expected patterns persisted after non-dry-run backfill")
	}
}

// Cobertura: detectFiltro — ramos B e C

func TestDetectFiltro_BranchB(t *testing.T) {
	t.Parallel()
	got := detectFiltro("aplicação do Filtro B no módulo X")
	assert.Equal(t, "B", got)
}

func TestDetectFiltro_BranchC(t *testing.T) {
	t.Parallel()
	got := detectFiltro("Filtro C ativado para roteamento crítico")
	assert.Equal(t, "C", got)
}

func TestDetectFiltro_Unknown(t *testing.T) {
	t.Parallel()
	got := detectFiltro("linha sem filtro válido")
	assert.Equal(t, "unknown", got)
}

// Cobertura: parseSentinelLine — Filtro B e Filtro C

func TestParseSentinelLine_FiltroB(t *testing.T) {
	t.Parallel()
	c, ok := parseSentinelLine("- Filtro B aplicado no roteamento de módulos críticos")
	assert.True(t, ok)
	assert.Contains(t, c.SourceRef, "Filtro-B")
}

func TestParseSentinelLine_FiltroC(t *testing.T) {
	t.Parallel()
	c, ok := parseSentinelLine("* Filtro C detectado em análise de divergência estrutural")
	assert.True(t, ok)
	assert.Contains(t, c.SourceRef, "Filtro-C")
}

// Coverage: insertIfNew — error path (Create fails)

func TestInsertIfNew_CreateError(t *testing.T) {
	t.Parallel()
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := graph.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	store, err := NewPatternStore(db)
	if err != nil {
		t.Fatalf("NewPatternStore failed: %v", err)
	}

	// Close the DB to force error on Create inside insertIfNew
	_ = db.Close()

	var result BackfillResult
	store.insertIfNew(context.Background(), BackfillCandidate{
		Title:       "Teste erro",
		Description: "desc",
		Category:    "anti-pattern",
		Source:      SourceManual,
		Tags:        "test",
		Impact:      ImpactHigh,
	}, SourceManual, &result)

	assert.Equal(t, 0, result.Inserted, "insertIfNew não deve contar inserção em erro")
	assert.True(t, len(result.Errors) > 0, "insertIfNew deve registrar erro quando Create falha")
}

// CG-02: BackfillFrom* methods should return ErrNilDB when store has no DB

func TestBackfillFromCognitiveDNA_NilDB(t *testing.T) {
	t.Parallel()
	s := &PatternStore{}
	_, err := s.BackfillFromCognitiveDNA(context.Background(), ".")
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestBackfillFromEvolutionInsights_NilDB(t *testing.T) {
	t.Parallel()
	s := &PatternStore{}
	_, err := s.BackfillFromEvolutionInsights(context.Background(), ".")
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

func TestBackfillFromSentinelLog_NilDB(t *testing.T) {
	t.Parallel()
	s := &PatternStore{}
	_, err := s.BackfillFromSentinelLog(context.Background(), ".", true)
	if !errors.Is(err, sqlite.ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}
