import type { RoundResult } from '../types'

interface ResultCardProps {
  result: RoundResult
  totalRounds: number
  onNext?: () => void
}

export function ResultCard({ result, totalRounds, onNext }: ResultCardProps) {
  const sorted = [...result.tally].sort((a, b) => b.count - a.count)

  return (
    <div style={{ animation: 'fadeIn 0.5s ease' }}>
      <h2 style={{ color: '#ffd700', margin: '0 0 0.5rem' }}>
        Round {result.round} / {totalRounds} — Result
      </h2>

      <p style={{ color: '#b0bec5', margin: '0 0 1.25rem', lineHeight: 1.6 }}>
        {result.narrativeOutcome}
      </p>

      {/* Tally bars */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1.5rem' }}>
        {sorted.map((t) => (
          <div key={t.optionId}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: '0.85rem',
                color: t.optionId === result.winningOption ? '#4caf50' : '#e0e0e0',
                marginBottom: '0.2rem',
              }}
            >
              <span>
                {t.optionId === result.winningOption && '🏆 '}
                {t.label || t.optionId}
              </span>
              <span>{t.count} vote{t.count !== 1 ? 's' : ''} ({t.pct.toFixed(0)}%)</span>
            </div>
            <div
              style={{
                height: 8,
                background: '#333',
                borderRadius: 4,
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  width: `${t.pct}%`,
                  background: t.optionId === result.winningOption ? '#4caf50' : '#607d8b',
                  transition: 'width 0.8s ease',
                  borderRadius: 4,
                }}
              />
            </div>
          </div>
        ))}
      </div>

      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          background: 'rgba(255,255,255,0.05)',
          borderRadius: 8,
          padding: '0.75rem 1rem',
          marginBottom: '1.25rem',
        }}
      >
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: '#aaa', textTransform: 'uppercase' }}>
            Consensus
          </div>
          <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#ffd700' }}>
            {result.consensusPct.toFixed(0)}%
          </div>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: '#aaa', textTransform: 'uppercase' }}>
            Peace Score
          </div>
          <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#4caf50' }}>
            +{result.peaceScore.toFixed(0)}
          </div>
        </div>
      </div>

      {onNext && (
        <button className="btn-primary" onClick={onNext}>
          Next Round →
        </button>
      )}
    </div>
  )
}
