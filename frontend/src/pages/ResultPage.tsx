import { useGameStore } from '../store'
import { ResultCard } from '../components/ResultCard'
import { GAME_SERVER_URL } from '../config'
import { useState } from 'react'

export function ResultPage() {
  const { gameState, sessionId, displayName } = useGameStore()
  const [advancing, setAdvancing] = useState(false)

  const result = gameState?.lastResult
  if (!result || !gameState) return null

  const isHost = gameState.players[0]?.displayName === displayName
  const hasMore = gameState.currentRound < gameState.totalRounds

  async function handleNext() {
    if (!sessionId) return
    setAdvancing(true)
    try {
      await fetch(`${GAME_SERVER_URL}/sessions/${sessionId}/start`, { method: 'POST' })
    } finally {
      setAdvancing(false)
    }
  }

  return (
    <div className="card" style={{ maxWidth: 560 }}>
      <ResultCard
        result={result}
        totalRounds={gameState.totalRounds}
        onNext={isHost && hasMore ? () => void handleNext() : undefined}
      />
      {advancing && (
        <p style={{ color: '#90a4ae', textAlign: 'center', marginTop: '0.75rem' }}>
          Loading next crisis…
        </p>
      )}
    </div>
  )
}
