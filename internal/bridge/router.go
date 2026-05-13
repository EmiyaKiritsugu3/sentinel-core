// Package bridge connects the agent engine to external AI providers.
package bridge

// ContextStrategy defines what context to inject into the AI payload
// based on the classified task intent.
type ContextStrategy struct {
	// DB-queryable fields (nodes + edges tables)
	HighCoupling  bool // nodes with highest fan-in via edges COUNT(*)
	RecentChanges bool // weight toward higher last_indexed
	IncludeTests  bool // nodes where file_path LIKE '%_test.go'
	NodeLimit     int  // max nodes to inject (0 = use current default)

	// File-based fields (direct filesystem read)
	IncludeADRs        bool // reads docs/architecture/adr/*.md
	IncludeDebtMarkers bool // reads docs/process/TECHNICAL-DEBT.md, filters by task keywords
}

var strategyByIntent = map[Intent]ContextStrategy{
	IntentDiagnose: {
		HighCoupling:  true,
		RecentChanges: true,
		NodeLimit:     15,
	},
	IntentImplement: {
		IncludeTests: true,
		IncludeADRs:  true,
		NodeLimit:    10,
	},
	IntentRefactor: {
		HighCoupling:       true,
		IncludeDebtMarkers: true,
		NodeLimit:          12,
	},
	IntentReview: {
		IncludeADRs: true,
		NodeLimit:   8,
	},
	IntentUnknown: {}, // zero value → Factory uses existing default behavior
}

// StrategyFor returns the ContextStrategy for a given intent.
// IntentUnknown returns a zero-value strategy (no routing).
func StrategyFor(intent Intent) ContextStrategy {
	return strategyByIntent[intent]
}
