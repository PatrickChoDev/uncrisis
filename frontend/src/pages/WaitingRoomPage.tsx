import { useGameStore } from '../store'
import { PlayerList } from '../components/PlayerList'
import { GAME_SERVER_URL } from '../config'
import { useState } from 'react'

export function WaitingRoomPage() {
  const { sessionId, gameState, displayName } = useGameStore()
  const [starting, setStarting] = useState(false)

  const players = gameState?.players ?? []
  const isHost = players[0]?.displayName === displayName

  async function handleStart() {
    if (!sessionId) return
    setStarting(true)
    try {
      await fetch(`${GAME_SERVER_URL}/sessions/${sessionId}/start`, {
        method: 'POST',
      })
    } finally {
      setStarting(false)
    }
  }

  return (
    <div className="card" style={{ maxWidth: 480 }}>
      <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
        <div style={{ fontSize: '2rem' }}>🏛️</div>
        <h2 style={{ margin: '0.5rem 0 0.25rem', color: '#ffd700' }}>Security Council Chamber</h2>
        <p style={{ color: '#90a4ae', margin: 0, fontSize: '0.85rem' }}>
          Room Code:{' '}
          <strong style={{ color: '#fff', letterSpacing: '0.1em' }}>{sessionId}</strong>
        </p>
      </div>

      <div style={{ marginBottom: '1.5rem' }}>
        <h3 style={{ margin: '0 0 0.75rem', fontSize: '0.9rem', color: '#90a4ae', textTransform: 'uppercase' }}>
          Delegates ({players.length})
        </h3>
        <PlayerList players={players} />
      </div>

      {isHost ? (
        <button
          className="btn-primary"
          onClick={() => void handleStart()}
          disabled={starting || players.length < 1}
        >
          {starting ? 'Starting…' : '▶ Start Crisis'}
        </button>
      ) : (
        <p style={{ color: '#90a4ae', textAlign: 'center', fontSize: '0.9rem', margin: 0 }}>
          Waiting for the host to start the session…
        </p>
      )}
    </div>
  )
}
