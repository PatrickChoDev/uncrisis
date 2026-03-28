// Package appsync implements the game.Notifier interface via the AppSync HTTP endpoint.
package appsync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PatrickChoDev/uncrisis/backend/internal/game"
)

const updateGameStateMutation = `
mutation UpdateGameState($input: GameStateInput!) {
  updateGameState(input: $input) {
    sessionId
    phase
    currentRound
    totalRounds
    timeLeftSecs
    peaceScore
  }
}
`

// Client posts GraphQL mutations to AppSync.
type Client struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

// New creates an AppSync client.
func New(endpoint, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BroadcastGameState sends an updateGameState mutation to AppSync,
// which triggers the onGameStateUpdated subscription for all connected players.
func (c *Client) BroadcastGameState(ctx context.Context, state game.GameState) error {
	input := gameStateToInput(state)

	payload := map[string]interface{}{
		"query": updateGameStateMutation,
		"variables": map[string]interface{}{
			"input": input,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("appsync marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("appsync request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("appsync http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("appsync returned status %d", resp.StatusCode)
	}

	var result struct {
		Errors []struct{ Message string } `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil // body might be empty on success
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("appsync graphql error: %s", result.Errors[0].Message)
	}

	return nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func gameStateToInput(s game.GameState) map[string]interface{} {
	m := map[string]interface{}{
		"sessionId":    s.SessionID,
		"phase":        string(s.Phase),
		"currentRound": s.CurrentRound,
		"totalRounds":  s.TotalRounds,
		"players":      playersToInput(s.Players),
		"timeLeftSecs": s.TimeLeftSecs,
		"peaceScore":   s.PeaceScore,
	}
	if s.Scenario != nil {
		m["scenario"] = scenarioToInput(s.Scenario)
	}
	if s.LastResult != nil {
		m["lastResult"] = resultToInput(s.LastResult)
	}
	return m
}

func playersToInput(players []game.Player) []map[string]interface{} {
	out := make([]map[string]interface{}, len(players))
	for i, p := range players {
		out[i] = map[string]interface{}{
			"playerId":    p.PlayerID,
			"displayName": p.DisplayName,
			"hasVoted":    p.HasVoted,
		}
	}
	return out
}

func scenarioToInput(s *game.Scenario) map[string]interface{} {
	opts := make([]map[string]string, len(s.Options))
	for i, o := range s.Options {
		opts[i] = map[string]string{"id": o.ID, "label": o.Label}
	}
	return map[string]interface{}{
		"scenarioId": s.ScenarioID,
		"title":      s.Title,
		"context":    s.Context,
		"options":    opts,
	}
}

func resultToInput(r *game.RoundResult) map[string]interface{} {
	tally := make([]map[string]interface{}, len(r.Tally))
	for i, t := range r.Tally {
		tally[i] = map[string]interface{}{
			"optionId": t.OptionID,
			"label":    t.Label,
			"count":    t.Count,
			"pct":      t.Pct,
		}
	}
	return map[string]interface{}{
		"round":            r.Round,
		"tally":            tally,
		"winningOption":    r.WinningOption,
		"consensusPct":     r.ConsensusPct,
		"peaceScore":       r.PeaceScore,
		"narrativeOutcome": r.NarrativeOutcome,
	}
}
