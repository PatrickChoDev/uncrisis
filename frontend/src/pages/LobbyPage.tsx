import { useState } from 'react'
import { v4 as uuidv4 } from 'uuid'
import { useGameStore } from '../store'
import { GAME_SERVER_URL } from '../config'

export function LobbyPage() {
  const [mode, setMode] = useState<'home' | 'create' | 'join'>('home')
  const [hostName, setHostName] = useState('')
  const [roomCode, setRoomCode] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const setSession = useGameStore((s) => s.setSession)

  async function handleCreate() {
    if (!hostName.trim()) { setError('Enter your display name'); return }
    setLoading(true)
    setError('')
    try {
      const res = await fetch(`${GAME_SERVER_URL}/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ hostName: hostName.trim(), totalRounds: 5 }),
      })
      if (!res.ok) throw new Error(await res.text())
      const { sessionId } = (await res.json()) as { sessionId: string }
      const playerId = uuidv4()
      setSession(sessionId, playerId, hostName.trim())
      // Join own session
      await fetch(`${GAME_SERVER_URL}/sessions/${sessionId}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ playerId, displayName: hostName.trim() }),
      })
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  async function handleJoin() {
    if (!roomCode.trim() || !displayName.trim()) {
      setError('Enter both room code and display name')
      return
    }
    setLoading(true)
    setError('')
    try {
      const playerId = uuidv4()
      const res = await fetch(`${GAME_SERVER_URL}/sessions/${roomCode.trim()}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ playerId, displayName: displayName.trim() }),
      })
      if (!res.ok) throw new Error(await res.text())
      setSession(roomCode.trim(), playerId, displayName.trim())
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card" style={{ maxWidth: 440 }}>
      {/* Logo / title */}
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <div style={{ fontSize: '3rem' }}>🌐</div>
        <h1 style={{ margin: '0.25rem 0 0.5rem', fontSize: '1.8rem', color: '#ffd700' }}>
          UN Crisis Room
        </h1>
        <p style={{ color: '#90a4ae', margin: 0, fontSize: '0.9rem' }}>
          Real-time multiplayer diplomacy game
        </p>
      </div>

      {mode === 'home' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <button className="btn-primary" onClick={() => setMode('create')}>
            🏠 Create Room
          </button>
          <button className="btn-secondary" onClick={() => setMode('join')}>
            🚪 Join Room
          </button>
        </div>
      )}

      {mode === 'create' && (
        <form
          onSubmit={(e) => { e.preventDefault(); void handleCreate() }}
          style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}
        >
          <label className="field-label">Your Display Name</label>
          <input
            className="text-input"
            placeholder="Ambassador Smith"
            value={hostName}
            onChange={(e) => setHostName(e.target.value)}
            autoFocus
          />
          {error && <p style={{ color: '#f44336', margin: 0, fontSize: '0.85rem' }}>{error}</p>}
          <button className="btn-primary" type="submit" disabled={loading}>
            {loading ? 'Creating…' : '🚀 Create & Enter'}
          </button>
          <button className="btn-ghost" type="button" onClick={() => { setMode('home'); setError('') }}>
            ← Back
          </button>
        </form>
      )}

      {mode === 'join' && (
        <form
          onSubmit={(e) => { e.preventDefault(); void handleJoin() }}
          style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}
        >
          <label className="field-label">Room Code</label>
          <input
            className="text-input"
            placeholder="e.g. a1b2c3d4"
            value={roomCode}
            onChange={(e) => setRoomCode(e.target.value)}
            autoFocus
          />
          <label className="field-label">Your Display Name</label>
          <input
            className="text-input"
            placeholder="Ambassador Jones"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
          />
          {error && <p style={{ color: '#f44336', margin: 0, fontSize: '0.85rem' }}>{error}</p>}
          <button className="btn-primary" type="submit" disabled={loading}>
            {loading ? 'Joining…' : '🤝 Join Session'}
          </button>
          <button className="btn-ghost" type="button" onClick={() => { setMode('home'); setError('') }}>
            ← Back
          </button>
        </form>
      )}
    </div>
  )
}
