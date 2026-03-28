import { useGameStore } from '../store'

export function FinalPage() {
  const { gameState, reset } = useGameStore()
  if (!gameState) return null

  const maxScore = gameState.totalRounds * 100
  const pct = Math.min((gameState.peaceScore / maxScore) * 100, 100)

  let medal = '🌱'
  let label = 'Emerging Diplomats'
  if (pct >= 80) { medal = '🏆'; label = 'Master Peacemakers' }
  else if (pct >= 60) { medal = '🕊️'; label = 'Skilled Negotiators' }
  else if (pct >= 40) { medal = '🤝'; label = 'Seasoned Diplomats' }

  return (
    <div className="card" style={{ maxWidth: 480, textAlign: 'center' }}>
      <div style={{ fontSize: '4rem', marginBottom: '0.5rem' }}>{medal}</div>
      <h1 style={{ color: '#ffd700', margin: '0 0 0.25rem' }}>{label}</h1>
      <p style={{ color: '#90a4ae', margin: '0 0 2rem', fontSize: '0.9rem' }}>
        The Security Council has reached its final deliberation.
      </p>

      {/* Peace score ring */}
      <div
        style={{
          position: 'relative',
          width: 140,
          height: 140,
          margin: '0 auto 1.5rem',
        }}
      >
        <svg width="140" height="140" style={{ transform: 'rotate(-90deg)' }}>
          <circle cx="70" cy="70" r="60" fill="none" stroke="#333" strokeWidth="12" />
          <circle
            cx="70"
            cy="70"
            r="60"
            fill="none"
            stroke="#4caf50"
            strokeWidth="12"
            strokeDasharray={`${2 * Math.PI * 60}`}
            strokeDashoffset={`${2 * Math.PI * 60 * (1 - pct / 100)}`}
            strokeLinecap="round"
            style={{ transition: 'stroke-dashoffset 1.5s ease' }}
          />
        </svg>
        <div
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <span style={{ fontSize: '1.75rem', fontWeight: 700, color: '#fff' }}>
            {pct.toFixed(0)}%
          </span>
          <span style={{ fontSize: '0.65rem', color: '#90a4ae', textTransform: 'uppercase' }}>
            Peace
          </span>
        </div>
      </div>

      {/* Per-round scores */}
      <div style={{ marginBottom: '1.5rem', textAlign: 'left' }}>
        {gameState.players.length > 0 && (
          <p style={{ color: '#e0e0e0', fontSize: '0.85rem', margin: '0 0 0.5rem' }}>
            <strong>{gameState.players.length}</strong> delegate
            {gameState.players.length !== 1 ? 's' : ''} participated across{' '}
            <strong>{gameState.totalRounds}</strong> rounds.
          </p>
        )}
        <p style={{ color: '#90a4ae', fontSize: '0.8rem', margin: 0 }}>
          Collective Peace Score: {gameState.peaceScore.toFixed(0)} / {maxScore}
        </p>
      </div>

      <button className="btn-primary" onClick={reset}>
        🔄 Play Again
      </button>
    </div>
  )
}
