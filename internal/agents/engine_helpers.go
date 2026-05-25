package agents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/knowledge"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/math"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/generative-ai-go/genai"
)

// readPriorTrust reads the agent's historical trust data from the DB.
// Returns (successes, total, trustScore, error). If no row exists, returns (0, 0, 0.5, nil).
func readPriorTrust(ctx context.Context, db *sqlite.DB, agentName string) (successes, total int, trust float64, err error) {
	if err := db.Conn.QueryRowContext(ctx,
		"SELECT successes, total FROM agent_trust WHERE agent_name = ?",
		agentName,
	).Scan(&successes, &total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, math.CalculateTrustScore(0, 0), nil
		}
		slog.Warn("failed to read prior trust", "agent", agentName, "error", err)
		return 0, 0, math.CalculateTrustScore(0, 0), err
	}
	return successes, total, math.CalculateTrustScore(successes, total), nil
}

// persistTrust updates the agent's trust record in the DB after execution completes.
func persistTrust(ctx context.Context, db *sqlite.DB, agentName string, priorSuccesses, priorTotal int, success bool) error {
	newTotal := priorTotal + 1
	newSuccesses := priorSuccesses
	if success {
		newSuccesses++
	}
	trustScore := math.CalculateTrustScore(newSuccesses, newTotal)
	if _, err := db.Conn.ExecContext(ctx,
		`INSERT INTO agent_trust (agent_name, successes, total, trust_score)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(agent_name) DO UPDATE SET
		successes = excluded.successes,
		total = excluded.total,
		trust_score = excluded.trust_score,
		updated_at = CURRENT_TIMESTAMP`,
		agentName, newSuccesses, newTotal, trustScore,
	); err != nil {
		slog.Warn("failed to persist trust", "agent", agentName, "successes", newSuccesses, "total", newTotal, "trust", trustScore, "error", err)
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
		slog.Warn("entropy threshold exceeded", "lambda", lambda, "max", effectiveMaxLambda)
		knowledge.GlobalBuffer.Record(knowledge.SessionEvent{
			Type:    knowledge.EventPattern,
			Domain:  "engine",
			Summary: fmt.Sprintf("Gate A: entropy threshold exceeded (λ=%.2f > max=%.2f)", lambda, effectiveMaxLambda),
			Tags:    []string{"gate-a", "entropy", "intervention"},
		})
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
			slog.Warn("logic drift detected", "divergence", divergence, "consecutive", newCount)
			knowledge.GlobalBuffer.Record(knowledge.SessionEvent{
				Type:    knowledge.EventPattern,
				Domain:  "engine",
				Summary: "Gate A.5: logic drift detected",
				Tags:    []string{"gate-a5", "divergence", "intervention"},
			})
			return newCount, true, msg
		}
		return newCount, false, ""
	}

	// Stable step: reset divergence count
	return 0, false, ""
}

// persistMetrics writes the final execution metrics (latency, tokens, cost, delta) to the tasks table.
func persistMetrics(ctx context.Context, db *sqlite.DB, stateID string, tokensUsed int, apiCost float64, latencyMs float64, priorTrust float64) error {
	probHallucination := 1.0 - priorTrust
	delta := math.CalculateDelta(probHallucination, 5.0, latencyMs, apiCost)

	query := "UPDATE tasks SET latency_ms = ?, tokens_used = ?, api_cost = ?, math_delta = ? WHERE id = ?"
	if _, err := db.Conn.ExecContext(ctx, query, latencyMs, tokensUsed, apiCost, delta, stateID); err != nil {
		slog.Warn("failed to persist math metrics", "task", stateID, "error", err)
		return err
	}
	return nil
}

// shouldTerminate returns true when the model response contains a Sovereign Audit
// report and no pending tool calls, indicating the agent has completed its task.
func shouldTerminate(toolCalls []map[string]interface{}, textResponses []string) bool {
	if len(toolCalls) == 0 {
		for _, text := range textResponses {
			if strings.Contains(strings.ToLower(text), "sovereign audit") {
				return true
			}
		}
	}
	return false
}
