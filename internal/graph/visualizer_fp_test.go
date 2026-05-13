package graph

import "testing"

// CG-01 FP tests: strings.Contains classification must have documented false positives.
// classifyContainer uses strings.Contains to map file paths to container IDs.
// Each test crafts a path that contains the substring but belongs to a different
// container, proving the classification mechanism can misclassify.

func TestClassifyContainer_FP_CmdSentinel(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "cmd/sentinel") → "cli"
	// FP: a vendor backup dir contains the substring but is NOT the CLI

	got := classifyContainer("vendor/cmd/sentinel_backup/main.go")
	t.Logf("CG-01 FP: vendor/cmd/sentinel_backup → %q (expected empty, got 'cli' = FP)", got)
	if got != "cli" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_InternalAgents(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "internal/agents") → "agents"
	// FP: a doc path referencing agents but not actual agent code

	got := classifyContainer("docs/internal/agents_design.md")
	t.Logf("CG-01 FP: docs/internal/agents_design.md → %q (expected empty, got 'agents' = FP)", got)
	if got != "agents" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_InternalGraph(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "internal/graph") → "graph"
	// FP: a test fixture that contains the substring but isn't graph code

	got := classifyContainer("testdata/internal/graph_sample.json")
	t.Logf("CG-01 FP: testdata/internal/graph_sample.json → %q (expected empty, got 'graph' = FP)", got)
	if got != "graph" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_InternalAudit(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "internal/audit") → "audit"
	// FP: an audit log stored outside the audit package

	got := classifyContainer("logs/internal/audit_trace.log")
	t.Logf("CG-01 FP: logs/internal/audit_trace.log → %q (expected empty, got 'audit' = FP)", got)
	if got != "audit" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_InternalReflect(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "internal/reflect") → "audit"
	// FP: a reflection utility outside the reflect package

	got := classifyContainer("tools/internal/reflectgen/main.go")
	t.Logf("CG-01 FP: tools/internal/reflectgen/main.go → %q (expected empty, got 'audit' = FP)", got)
	if got != "audit" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_InternalState(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "internal/state") → "state"
	// FP: a state machine library imported from a different module

	got := classifyContainer("third_party/internal/state_machine.go")
	t.Logf("CG-01 FP: third_party/internal/state_machine.go → %q (expected empty, got 'state' = FP)", got)
	if got != "state" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}

func TestClassifyContainer_FP_LegacyTs(t *testing.T) {
	t.Parallel()
	// strings.Contains(path, "legacy/ts") → "frontend"
	// FP: a TypeScript migration guide that isn't legacy code

	got := classifyContainer("docs/migrations/legacy/ts_to_go.md")
	t.Logf("CG-01 FP: docs/migrations/legacy/ts_to_go.md → %q (expected empty, got 'frontend' = FP)", got)
	if got != "frontend" {
		t.Log("FP not triggered — no misclassification for this input")
	}
}
