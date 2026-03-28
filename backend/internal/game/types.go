// Package game defines the core data types for the UN Crisis Room game.
package game

import "time"

// Phase represents the current state of a game session.
type Phase string

const (
	PhaseLobby  Phase = "LOBBY"
	PhaseCrisis Phase = "CRISIS"
	PhaseVoting Phase = "VOTING"
	PhaseTally  Phase = "TALLY"
	PhaseResult Phase = "RESULT"
	PhaseFinal  Phase = "FINAL"
)

// Option is one of the four resolution choices for a crisis scenario.
type Option struct {
	ID    string `json:"id" dynamodbav:"id"`
	Label string `json:"label" dynamodbav:"label"`
}

// Scenario holds a single crisis card.
type Scenario struct {
	ScenarioID string   `json:"scenarioId" dynamodbav:"scenarioId"`
	Title      string   `json:"title" dynamodbav:"title"`
	Context    string   `json:"context" dynamodbav:"context"`
	Options    []Option `json:"options" dynamodbav:"options"`
}

// Player represents a connected participant.
type Player struct {
	PlayerID    string `json:"playerId" dynamodbav:"playerId"`
	DisplayName string `json:"displayName" dynamodbav:"displayName"`
	HasVoted    bool   `json:"hasVoted" dynamodbav:"hasVoted"`
}

// OptionTally summarises the vote count for one option in a round.
type OptionTally struct {
	OptionID string  `json:"optionId" dynamodbav:"optionId"`
	Label    string  `json:"label" dynamodbav:"label"`
	Count    int     `json:"count" dynamodbav:"count"`
	Pct      float64 `json:"pct" dynamodbav:"pct"`
}

// RoundResult is the outcome of one voting round.
type RoundResult struct {
	Round            int           `json:"round" dynamodbav:"round"`
	Tally            []OptionTally `json:"tally" dynamodbav:"tally"`
	WinningOption    string        `json:"winningOption" dynamodbav:"winningOption"`
	ConsensusPct     float64       `json:"consensusPct" dynamodbav:"consensusPct"`
	PeaceScore       float64       `json:"peaceScore" dynamodbav:"peaceScore"`
	NarrativeOutcome string        `json:"narrativeOutcome" dynamodbav:"narrativeOutcome"`
}

// Session is the full in-memory + persisted state of a game room.
type Session struct {
	SessionID    string        `json:"sessionId" dynamodbav:"sessionId"`
	HostName     string        `json:"hostName" dynamodbav:"hostName"`
	Players      []Player      `json:"players" dynamodbav:"players"`
	Phase        Phase         `json:"phase" dynamodbav:"phase"`
	CurrentRound int           `json:"currentRound" dynamodbav:"currentRound"`
	TotalRounds  int           `json:"totalRounds" dynamodbav:"totalRounds"`
	Scenario     *Scenario     `json:"scenario,omitempty" dynamodbav:"scenario,omitempty"`
	Results      []RoundResult `json:"results,omitempty" dynamodbav:"results,omitempty"`
	PeaceScore   float64       `json:"peaceScore" dynamodbav:"peaceScore"`
	CreatedAt    time.Time     `json:"createdAt" dynamodbav:"createdAt"`
	ExpiresAt    int64         `json:"expiresAt" dynamodbav:"expiresAt"`
}

// Vote is a single player vote received from SQS.
type Vote struct {
	SessionID string `json:"sessionId"`
	PlayerID  string `json:"playerId"`
	Choice    string `json:"choice"`
	Round     int    `json:"round"`
}

// GameState is the payload broadcast to AppSync subscribers.
type GameState struct {
	SessionID    string       `json:"sessionId"`
	Phase        Phase        `json:"phase"`
	CurrentRound int          `json:"currentRound"`
	TotalRounds  int          `json:"totalRounds"`
	Scenario     *Scenario    `json:"scenario,omitempty"`
	Players      []Player     `json:"players"`
	TimeLeftSecs int          `json:"timeLeftSecs"`
	LastResult   *RoundResult `json:"lastResult,omitempty"`
	PeaceScore   float64      `json:"peaceScore"`
}

// NarrativeOutcomes maps winning options to flavour text.
// The key format is "<scenarioId>:<optionId>".
var NarrativeOutcomes = map[string]string{
	"default:A": "The ceasefire holds. Aid corridors open across the border.",
	"default:B": "Diplomatic back-channels restore dialogue after tense negotiations.",
	"default:C": "Sanctions bite; both parties agree to international mediation.",
	"default:D": "The Security Council passes a binding resolution — a rare moment of unity.",
}

// DefaultNarrative returns a generic narrative when no specific one is defined.
func DefaultNarrative(optionID string) string {
	if n, ok := NarrativeOutcomes["default:"+optionID]; ok {
		return n
	}
	return "The international community applauds the collective decision."
}
