// Package game implements the core game-loop logic for the UN Crisis Room.
package game

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager owns all active sessions and drives the round lifecycle.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState

	persist  Persister
	notifier Notifier
	logger   *slog.Logger

	roundDuration time.Duration
	totalRounds   int
}

// sessionState wraps Session with in-memory vote tracking.
type sessionState struct {
	mu          sync.Mutex
	session     *Session
	votes       map[string]string // playerId → choice
	timeLeftSecs int
	roundTimer  *time.Timer
	cancelRound context.CancelFunc
}

// Persister abstracts DynamoDB operations.
type Persister interface {
	SaveSession(ctx context.Context, s *Session) error
	GetScenario(ctx context.Context, idx int) (*Scenario, error)
	ScenarioCount(ctx context.Context) (int, error)
}

// Notifier abstracts AppSync mutations.
type Notifier interface {
	BroadcastGameState(ctx context.Context, state GameState) error
}

// NewManager creates a Manager.
func NewManager(p Persister, n Notifier, roundDuration time.Duration, totalRounds int, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		sessions:      make(map[string]*sessionState),
		persist:       p,
		notifier:      n,
		logger:        logger,
		roundDuration: roundDuration,
		totalRounds:   totalRounds,
	}
}

// CreateSession initialises a new game room and returns its sessionId.
func (m *Manager) CreateSession(ctx context.Context, hostName string) (string, error) {
	sessionID := uuid.New().String()[:8] // short code for sharing

	sess := &Session{
		SessionID:   sessionID,
		HostName:    hostName,
		Phase:       PhaseLobby,
		TotalRounds: m.totalRounds,
		Players:     []Player{},
		Results:     []RoundResult{},
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
	}

	if err := m.persist.SaveSession(ctx, sess); err != nil {
		return "", err
	}

	m.mu.Lock()
	m.sessions[sessionID] = &sessionState{
		session: sess,
		votes:   make(map[string]string),
	}
	m.mu.Unlock()

	return sessionID, nil
}

// JoinSession adds a player to an existing lobby and broadcasts the updated state.
func (m *Manager) JoinSession(ctx context.Context, sessionID, playerID, displayName string) error {
	ss, err := m.getState(sessionID)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	// Deduplicate — but still broadcast so late-connecting subscribers get current state
	for _, p := range ss.session.Players {
		if p.PlayerID == playerID {
			ss.mu.Unlock()
			return m.broadcastState(ctx, ss)
		}
	}
	ss.session.Players = append(ss.session.Players, Player{
		PlayerID:    playerID,
		DisplayName: displayName,
		HasVoted:    false,
	})
	ss.mu.Unlock()

	return m.broadcastState(ctx, ss)
}

// StartRound advances the session to the CRISIS/VOTING phase and begins the timer.
func (m *Manager) StartRound(ctx context.Context, sessionID string) error {
	ss, err := m.getState(sessionID)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	s := ss.session
	s.CurrentRound++

	// Pick scenario (cycle through scenarios if needed)
	scenarioIdx := (s.CurrentRound - 1) % len(DefaultScenarios)
	scenario := DefaultScenarios[scenarioIdx]
	s.Scenario = &scenario
	s.Phase = PhaseCrisis

	// Reset vote tracking
	ss.votes = make(map[string]string)
	for i := range s.Players {
		s.Players[i].HasVoted = false
	}
	ss.mu.Unlock()

	if err := m.persist.SaveSession(ctx, s); err != nil {
		return err
	}
	if err := m.broadcastState(ctx, ss); err != nil {
		return err
	}

	// Transition to VOTING after a brief "read the scenario" delay
	time.AfterFunc(3*time.Second, func() {
		ss.mu.Lock()
		ss.session.Phase = PhaseVoting
		ss.mu.Unlock()
		_ = m.broadcastState(context.Background(), ss)
	})

	// Start the round timer
	roundCtx, cancel := context.WithCancel(ctx)
	ss.mu.Lock()
	if ss.cancelRound != nil {
		ss.cancelRound()
	}
	ss.cancelRound = cancel
	ss.mu.Unlock()

	go m.runRoundTimer(roundCtx, ss)
	return nil
}

// ProcessVote records a player's vote received from SQS.
func (m *Manager) ProcessVote(ctx context.Context, vote Vote) error {
	ss, err := m.getState(vote.SessionID)
	if err != nil {
		m.logger.Warn("vote for unknown session", "sessionId", vote.SessionID)
		return nil
	}

	ss.mu.Lock()
	s := ss.session
	if s.Phase != PhaseVoting || vote.Round != s.CurrentRound {
		ss.mu.Unlock()
		return nil
	}

	ss.votes[vote.PlayerID] = vote.Choice
	for i, p := range s.Players {
		if p.PlayerID == vote.PlayerID {
			s.Players[i].HasVoted = true
			break
		}
	}

	allVoted := len(ss.votes) >= len(s.Players) && len(s.Players) > 0
	ss.mu.Unlock()

	if err := m.broadcastState(ctx, ss); err != nil {
		return err
	}

	if allVoted {
		return m.TallyRound(ctx, vote.SessionID)
	}
	return nil
}

// TallyRound finalises the round: computes results, persists them, and broadcasts.
func (m *Manager) TallyRound(ctx context.Context, sessionID string) error {
	ss, err := m.getState(sessionID)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	s := ss.session
	s.Phase = PhaseTally

	if ss.cancelRound != nil {
		ss.cancelRound()
	}

	// Count votes
	counts := make(map[string]int)
	for _, choice := range ss.votes {
		counts[choice]++
	}

	total := len(ss.votes)
	if total == 0 {
		total = 1 // avoid division by zero when no votes
	}

	// Find the scenario options for labels
	optionLabels := make(map[string]string)
	if s.Scenario != nil {
		for _, o := range s.Scenario.Options {
			optionLabels[o.ID] = o.Label
		}
	}

	var tally []OptionTally
	var winner string
	var maxCount int

	for optID, cnt := range counts {
		tally = append(tally, OptionTally{
			OptionID: optID,
			Label:    optionLabels[optID],
			Count:    cnt,
			Pct:      float64(cnt) / float64(total) * 100,
		})
		if cnt > maxCount {
			maxCount = cnt
			winner = optID
		}
	}

	// Sort tally by option id for deterministic output
	sort.Slice(tally, func(i, j int) bool { return tally[i].OptionID < tally[j].OptionID })

	consensusPct := float64(maxCount) / float64(total) * 100

	// Peace score: higher consensus → higher peace score (max 100 per round)
	peaceScore := consensusPct * 0.8
	if consensusPct >= 75 {
		peaceScore = 100
	}
	s.PeaceScore += peaceScore

	narrative := DefaultNarrative(winner)
	if s.Scenario != nil {
		if n, ok := NarrativeOutcomes[s.Scenario.ScenarioID+":"+winner]; ok {
			narrative = n
		}
	}

	result := RoundResult{
		Round:            s.CurrentRound,
		Tally:            tally,
		WinningOption:    winner,
		ConsensusPct:     consensusPct,
		PeaceScore:       peaceScore,
		NarrativeOutcome: narrative,
	}
	s.Results = append(s.Results, result)
	s.Phase = PhaseResult
	ss.mu.Unlock()

	if err := m.persist.SaveSession(ctx, s); err != nil {
		return err
	}

	if err := m.broadcastState(ctx, ss); err != nil {
		return err
	}

	// Check if the game is over
	if s.CurrentRound >= s.TotalRounds {
		time.AfterFunc(5*time.Second, func() {
			ss.mu.Lock()
			ss.session.Phase = PhaseFinal
			ss.mu.Unlock()
			_ = m.broadcastState(context.Background(), ss)
		})
	}

	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (m *Manager) getState(sessionID string) (*sessionState, error) {
	m.mu.RLock()
	ss, ok := m.sessions[sessionID]
	m.mu.RUnlock()
	if !ok {
		return nil, errUnknownSession(sessionID)
	}
	return ss, nil
}

func (m *Manager) broadcastState(ctx context.Context, ss *sessionState) error {
	ss.mu.Lock()
	s := ss.session
	timeLeft := ss.timeLeftSecs
	ss.mu.Unlock()

	var lastResult *RoundResult
	if len(s.Results) > 0 {
		r := s.Results[len(s.Results)-1]
		lastResult = &r
	}

	state := GameState{
		SessionID:    s.SessionID,
		Phase:        s.Phase,
		CurrentRound: s.CurrentRound,
		TotalRounds:  s.TotalRounds,
		Scenario:     s.Scenario,
		Players:      s.Players,
		TimeLeftSecs: timeLeft,
		LastResult:   lastResult,
		PeaceScore:   s.PeaceScore,
	}

	return m.notifier.BroadcastGameState(ctx, state)
}

func (m *Manager) runRoundTimer(ctx context.Context, ss *sessionState) {
	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			elapsed := t.Sub(start)
			remaining := m.roundDuration - elapsed
			if remaining < 0 {
				remaining = 0
			}

			ss.mu.Lock()
			if ss.session.Phase == PhaseVoting {
				ss.timeLeftSecs = int(remaining.Seconds())
			}
			ss.mu.Unlock()

			// Broadcast updated countdown to subscribers
			if ss.session.Phase == PhaseVoting {
				_ = m.broadcastState(ctx, ss)
			}

			if elapsed >= m.roundDuration {
				_ = m.TallyRound(ctx, ss.session.SessionID)
				return
			}
		}
	}
}

// errUnknownSession is a lightweight error type.
type errUnknownSession string

func (e errUnknownSession) Error() string {
	return "unknown session: " + string(e)
}
