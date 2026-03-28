import type { Player } from '../types'

interface PlayerListProps {
  players: Player[]
}

export function PlayerList({ players }: PlayerListProps) {
  if (players.length === 0) {
    return <p style={{ color: '#aaa', fontStyle: 'italic' }}>Waiting for players to join…</p>
  }
  return (
    <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
      {players.map((p) => (
        <li
          key={p.playerId}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.35rem 0',
            color: '#e0e0e0',
          }}
        >
          <span
            style={{
              width: 10,
              height: 10,
              borderRadius: '50%',
              background: p.hasVoted ? '#4caf50' : '#9e9e9e',
              display: 'inline-block',
              flexShrink: 0,
            }}
          />
          {p.displayName}
          {p.hasVoted && (
            <span style={{ fontSize: '0.75rem', color: '#4caf50', marginLeft: 'auto' }}>
              voted ✓
            </span>
          )}
        </li>
      ))}
    </ul>
  )
}
