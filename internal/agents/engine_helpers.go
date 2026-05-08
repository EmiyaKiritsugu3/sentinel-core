package agents

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
)

// readPriorTrust reads the agent's historical trust data from the DB.
// Returns (successes, total, trustScore, error). If no row exists, returns (0, 0, 0.5, nil).
func readPriorTrust(db *sqlite.DB, agentName string) (successes, total int, trust float64, err error) {
	if err := db.Conn.QueryRow(
		"SELECT successes, total FROM agent_trust WHERE agent_name = ?",
		agentName,
	).Scan(&successes, &total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, math.CalculateTrustScore(0, 0), nil
		}
		log.Printf("[SENTINEL] Warning: failed to read prior trust for '%s': %v", agentName, err)
		return 0, 0, math.CalculateTrustScore(0, 0), err
	}
	return successes, total, math.CalculateTrustScore(successes, total), nil
}

// persistTrust updates the agent's trust record in the DB after execution completes.
func persistTrust(db *sqlite.DB, agentName string, priorSuccesses, priorTotal int, success bool) error {
	newTotal := priorTotal + 1
	newSuccesses := priorSuccesses
	if success {
		newSuccesses++
	}
	trustScore := math.CalculateTrustScore(newSuccesses, newTotal)
	if _, err := db.Conn.Exec(
		`INSERT INTO agent_trust (agent_name, successes, total, trust_score)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(agent_name) DO UPDATE SET
		successes = excluded.successes,
		total = excluded.total,
		trust_score = excluded.trust_score,
		updated_at = CURRENT_TIMESTAMP`,
		agentName, newSuccesses, newTotal, trustScore,
	); err != nil {
		log.Printf("[SENTINEL] Warning: failed to persist trust for '%s' (successes=%d total=%d trust=%.4f): %v", agentName, newSuccesses, newTotal, trustScore, err)
		return err
	}
	return nil
}

// countThoughtActionTokens classifies response parts into thought vs action tokens.
// Returns (actionTokens, thoughtTokens) as approximate token counts (1 token ≈ 4 chars).
func countThoughtActionTokens(parts []genai.Part) (actionTokens, thoughtTokens int) {
	actionChars := 0
	thoughtChars := 0

	for _, part := range parts {
		if text, ok := part.(genai.Text); ok {
			sText := string(text)
			if isExplicitThoughtBlock(sText) {
				thoughtChars += len(sText)
			} else {
				actionChars += len(sText)
			}
		}
	}

	return actionChars / 4, thoughtChars / 4
}

// checkGateA evaluates whether the cumulative lambda exceeds the effective threshold.
// Returns (intervene=true, message) if the gate should fire; otherwise (false, "").
func checkGateA(lambda, effectiveMaxLambda float64) (intervene bool, message string) {
	if lambda > effectiveMaxLambda {
		msg := fmt.Sprintf("GATE A INTERVENTION: Your action-to-thought ratio is too high (%.2f). You are hallucinating excessive code without planning. Re-evaluate your strategy and output a detailed thought process before proceeding.", lambda)
		log.Printf("[GATE A] Entropy threshold exceeded (λ=%.2f > EffectiveMax=%.2f). Interrupting execution.", lambda, effectiveMaxLambda)
		return true, msg
	}
	return false, ""
}

// checkGateA5 evaluates per-step Lyapunov divergence for logic drift detection.
// Returns (newDivergenceCount, intervene, message).
// If intervene is true, the caller should send the intervention message and continue.
func checkGateA5(stepLambda, previousLambda float64, divergenceCount int) (newCount int, intervene bool, message string) {
	if previousLambda <= 0 {
		return divergenceCount, false, ""
	}

	divergence := math.CalculateDivergence(stepLambda, previousLambda)
	const divergenceThreshold = 1.0
	if divergence > divergenceThreshold {
		newCount = divergenceCount + 1
		if newCount >= 2 {
			msg := fmt.Sprintf(
				"GATE A.5 INTERVENTION: Logic Drift detected. Your reasoning trajectory is diverging (Δλ=%.2f). Stop and re-plan from scratch before generating more code.",
				divergence,
			)
			log.Printf("[GATE A.5] Logic Drift detected (divergence=%.2f, consecutive=%d). Interrupting.", divergence, newCount)
			return newCount, true, msg
		}
		return newCount, false, ""
	}

	// Stable step: reset divergence count
	return 0, false, ""
}

// persistMetrics writes the final execution metrics (latency, tokens, cost, delta) to the tasks table.
func persistMetrics(db *sqlite.DB, stateID string, tokensUsed int, apiCost float64, latencyMs float64, priorTrust float64) error {
	probHallucination := 1.0 - priorTrust
	delta := math.CalculateDelta(probHallucination, 5.0, latencyMs, apiCost)

	query := "UPDATE tasks SET latency_ms = ?, tokens_used = ?, api_cost = ?, math_delta = ? WHERE id = ?"
	if _, err := db.Conn.Exec(query, latencyMs, tokensUsed, apiCost, delta, stateID); err != nil {
		log.Printf("[SENTINEL] Warning: failed to persist math metrics for task %s: %v", stateID, err)
		return err
	}
	return nil
}

// containsSovereignAudit checks whether any text response contains the Sovereign Audit Report marker.
func containsSovereignAudit(textResponses []string) bool {
	return strings.Contains(strings.Join(textResponses, ""), "Sovereign Audit Report")
}
