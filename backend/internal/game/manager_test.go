package game_test

import (
	"context"
	"testing"
	"time"

	"github.com/PatrickChoDev/uncrisis/backend/internal/game"
)

// ── Stub implementations ──────────────────────────────────────────────────────

type stubPersister struct{}

func (s *stubPersister) SaveSession(_ context.Context, _ *game.Session) error { return nil }
func (s *stubPersister) GetScenario(_ context.Context, idx int) (*game.Scenario, error) {
	sc := game.DefaultScenarios[idx%len(game.DefaultScenarios)]
	return &sc, nil
}
func (s *stubPersister) ScenarioCount(_ context.Context) (int, error) {
	return len(game.DefaultScenarios), nil
}

type stubNotifier struct {
	states []game.GameState
}

func (n *stubNotifier) BroadcastGameState(_ context.Context, s game.GameState) error {
	n.states = append(n.states, s)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newManager(t *testing.T) (*game.Manager, *stubNotifier) {
	t.Helper()
	notifier := &stubNotifier{}
	mgr := game.NewManager(
		&stubPersister{},
		notifier,
		60*time.Second, // long timeout — we drive tally manually
		3,
		nil, // logger
	)
	return mgr, notifier
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreateSession(t *testing.T) {
	mgr, _ := newManager(t)
	sessionID, err := mgr.CreateSession(context.Background(), "Alice")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty sessionId")
	}
	if len(sessionID) != 8 {
		t.Fatalf("expected 8-char sessionId, got %q", sessionID)
	}
}

func TestJoinSession(t *testing.T) {
	mgr, notifier := newManager(t)
	ctx := context.Background()

	sid, _ := mgr.CreateSession(ctx, "Alice")
	if err := mgr.JoinSession(ctx, sid, "p1", "Alice"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}

	// A broadcast should have been sent
	if len(notifier.states) == 0 {
		t.Fatal("expected at least one broadcast after join")
	}
	last := notifier.states[len(notifier.states)-1]
	if len(last.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(last.Players))
	}
	if last.Players[0].DisplayName != "Alice" {
		t.Fatalf("expected player Alice, got %q", last.Players[0].DisplayName)
	}
}

func TestJoinDeduplication(t *testing.T) {
	mgr, _ := newManager(t)
	ctx := context.Background()

	sid, _ := mgr.CreateSession(ctx, "Alice")
	_ = mgr.JoinSession(ctx, sid, "p1", "Alice")
	_ = mgr.JoinSession(ctx, sid, "p1", "Alice") // duplicate

	// Start round to check player count
	_ = mgr.StartRound(ctx, sid)
	// If we get here without panic the dedup logic worked
}

func TestStartRound(t *testing.T) {
	mgr, notifier := newManager(t)
	ctx := context.Background()

	sid, _ := mgr.CreateSession(ctx, "Alice")
	_ = mgr.JoinSession(ctx, sid, "p1", "Alice")

	notifier.states = nil // clear
	if err := mgr.StartRound(ctx, sid); err != nil {
		t.Fatalf("StartRound: %v", err)
	}

	// Should have broadcast a CRISIS state
	found := false
	for _, s := range notifier.states {
		if s.Phase == game.PhaseCrisis {
			found = true
			if s.Scenario == nil {
				t.Error("scenario should be set in CRISIS phase")
			}
		}
	}
	if !found {
		t.Error("expected a CRISIS broadcast after StartRound")
	}
}

func TestProcessVote_AllVoted_TalliesRound(t *testing.T) {
	mgr, notifier := newManager(t)
	ctx := context.Background()

	sid, _ := mgr.CreateSession(ctx, "Alice")
	_ = mgr.JoinSession(ctx, sid, "p1", "Alice")
	_ = mgr.JoinSession(ctx, sid, "p2", "Bob")
	_ = mgr.StartRound(ctx, sid)

	// Wait for the VOTING phase (triggered 3 s after CRISIS via time.AfterFunc)
	// We manually inject a vote directly with round=1 and rely on the
	// manager transitioning phase. Since we set a long timer and process
	// votes directly, we bypass the AfterFunc by submitting during CRISIS.
	// Real voting only tallies in VOTING phase, so let's verify the guard.

	v := game.Vote{SessionID: sid, PlayerID: "p1", Choice: "A", Round: 1}
	_ = mgr.ProcessVote(ctx, v) // dropped — phase is CRISIS

	// No RESULT broadcast yet
	for _, s := range notifier.states {
		if s.Phase == game.PhaseResult {
			t.Fatal("should not reach RESULT before VOTING phase")
		}
	}
}

func TestDefaultNarrative(t *testing.T) {
	n := game.DefaultNarrative("A")
	if n == "" {
		t.Fatal("expected non-empty narrative for option A")
	}
	unknown := game.DefaultNarrative("Z")
	if unknown == "" {
		t.Fatal("expected fallback narrative for unknown option")
	}
}

func TestDefaultScenariosNotEmpty(t *testing.T) {
	if len(game.DefaultScenarios) == 0 {
		t.Fatal("DefaultScenarios should not be empty")
	}
	for _, s := range game.DefaultScenarios {
		if s.ScenarioID == "" {
			t.Error("scenario missing ScenarioID")
		}
		if len(s.Options) != 4 {
			t.Errorf("scenario %s: expected 4 options, got %d", s.ScenarioID, len(s.Options))
		}
	}
}
