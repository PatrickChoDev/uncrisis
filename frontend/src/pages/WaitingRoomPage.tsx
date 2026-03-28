import { useEffect, useState } from 'react'
import { useGameStore } from '../store'
import { PlayerList } from '../components/PlayerList'
import { GAME_SERVER_URL } from '../config'

export function WaitingRoomPage() {
  const { sessionId, playerId, gameState, displayName } = useGameStore()
  const [starting, setStarting] = useState(false)

  const players = gameState?.players ?? []
  // Use playerId (not displayName) so two players with the same name don't both get host controls
  const isHost = !!playerId && players[0]?.playerId === playerId

  // Re-join on mount so the server broadcasts current state to our (now-ready) subscription.
  // This handles the race where the initial join broadcast fired before the WebSocket was open.
  useEffect(() => {
    if (!sessionId || !playerId || !displayName) return
    fetch(`${GAME_SERVER_URL}/sessions/${sessionId}/join`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ playerId, displayName }),
    }).catch(() => {/* best-effort refresh */})
  }, [sessionId, playerId, displayName])

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
