import { useGameStore } from '../store'
import { Timer } from '../components/Timer'
import { submitVote } from '../graphql'

export function VotingPage() {
  const { gameState, playerId, sessionId, selectedOption, setSelectedOption } = useGameStore()
  const scenario = gameState?.scenario
  const phase = gameState?.phase
  const myPlayer = gameState?.players.find((p) => p.playerId === playerId)

  async function handleVote(optionId: string) {
    if (!sessionId || !playerId || myPlayer?.hasVoted) return
    setSelectedOption(optionId)
    try {
      await submitVote(sessionId, playerId, optionId, gameState?.currentRound ?? 1)
    } catch (e) {
      console.error('vote error', e)
      setSelectedOption(null)
    }
  }

  if (!scenario) {
    return (
      <div className="card" style={{ textAlign: 'center' }}>
        <div style={{ fontSize: '2rem' }}>⚡</div>
        <h2 style={{ color: '#ffd700' }}>Crisis Briefing</h2>
        <p style={{ color: '#90a4ae' }}>Loading scenario…</p>
      </div>
    )
  }

  const isVoting = phase === 'VOTING'
  const isCrisis = phase === 'CRISIS'

  return (
    <div className="card" style={{ maxWidth: 600 }}>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1.25rem',
        }}
      >
        <div>
          <span style={{ fontSize: '0.75rem', color: '#90a4ae', textTransform: 'uppercase' }}>
            Round {gameState?.currentRound} / {gameState?.totalRounds}
          </span>
          <h2 style={{ margin: '0.25rem 0 0', color: '#ffd700', fontSize: '1.3rem' }}>
            {scenario.title}
          </h2>
        </div>
        {isVoting && <Timer secondsLeft={gameState?.timeLeftSecs ?? 30} />}
      </div>

      {/* Scenario context */}
      <p
        style={{
          color: '#cfd8dc',
          lineHeight: 1.7,
          background: 'rgba(255,255,255,0.04)',
          borderRadius: 8,
          padding: '1rem',
          marginBottom: '1.5rem',
          fontSize: '0.9rem',
        }}
      >
        {scenario.context}
      </p>

      {isCrisis && (
        <p style={{ color: '#ffd700', textAlign: 'center', animation: 'pulse 1s infinite' }}>
          ⏳ Delegates, please study the briefing…
        </p>
      )}

      {/* Vote options */}
      {isVoting && (
        <>
          <h3
            style={{
              margin: '0 0 0.75rem',
              fontSize: '0.85rem',
              color: '#90a4ae',
              textTransform: 'uppercase',
            }}
          >
            {myPlayer?.hasVoted ? 'Your vote has been cast' : 'Select a resolution:'}
          </h3>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem' }}>
            {scenario.options.map((opt) => {
              const isSelected = selectedOption === opt.id
              const voted = myPlayer?.hasVoted

              return (
                <button
                  key={opt.id}
                  onClick={() => void handleVote(opt.id)}
                  disabled={!!voted}
                  style={{
                    textAlign: 'left',
                    padding: '0.85rem 1rem',
                    borderRadius: 8,
                    border: isSelected ? '2px solid #4caf50' : '2px solid rgba(255,255,255,0.1)',
                    background: isSelected
                      ? 'rgba(76,175,80,0.15)'
                      : 'rgba(255,255,255,0.04)',
                    color: '#e0e0e0',
                    cursor: voted ? 'default' : 'pointer',
                    fontSize: '0.9rem',
                    transition: 'all 0.2s',
                    opacity: voted && !isSelected ? 0.5 : 1,
                  }}
                >
                  <span
                    style={{
                      fontWeight: 700,
                      color: isSelected ? '#4caf50' : '#ffd700',
                      marginRight: '0.5rem',
                    }}
                  >
                    {opt.id}.
                  </span>
                  {opt.label}
                  {isSelected && (
                    <span style={{ float: 'right', color: '#4caf50' }}>✓</span>
                  )}
                </button>
              )
            })}
          </div>
        </>
      )}

      {/* Players who have voted */}
      {isVoting && (
        <div style={{ marginTop: '1.5rem' }}>
          <p
            style={{
              fontSize: '0.75rem',
              color: '#90a4ae',
              textTransform: 'uppercase',
              margin: '0 0 0.5rem',
            }}
          >
            Votes cast:{' '}
            {gameState?.players.filter((p) => p.hasVoted).length ?? 0} /{' '}
            {gameState?.players.length ?? 0}
          </p>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            {gameState?.players.map((p) => (
              <span
                key={p.playerId}
                style={{
                  padding: '0.2rem 0.6rem',
                  borderRadius: 12,
                  fontSize: '0.75rem',
                  background: p.hasVoted ? 'rgba(76,175,80,0.2)' : 'rgba(255,255,255,0.06)',
                  color: p.hasVoted ? '#4caf50' : '#90a4ae',
                }}
              >
                {p.displayName}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
