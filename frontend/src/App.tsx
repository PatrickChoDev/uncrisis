import { useEffect } from 'react'
import { Amplify } from 'aws-amplify'
import { amplifyConfig } from './config'
import { useGameStore } from './store'
import { subscribeToGameState } from './graphql'
import type { GameState } from './types'

import { GlobeCanvas } from './components/GlobeCanvas'
import { LobbyPage } from './pages/LobbyPage'
import { WaitingRoomPage } from './pages/WaitingRoomPage'
import { VotingPage } from './pages/VotingPage'
import { ResultPage } from './pages/ResultPage'
import { FinalPage } from './pages/FinalPage'

Amplify.configure(amplifyConfig)

export default function App() {
  const { sessionId, gameState, setGameState, setSelectedOption } = useGameStore()
  const phase = gameState?.phase

  // Subscribe to real-time game state updates whenever we have a session
  useEffect(() => {
    if (!sessionId) return

    const sub = subscribeToGameState(
      sessionId,
      (data) => {
        const state = data as GameState
        if (state) {
          setGameState(state)
          // Reset selected option when a new round starts
          if (state.phase === 'VOTING') {
            setSelectedOption(null)
          }
        }
      },
      (err) => console.error('subscription error', err),
    )

    return () => {
      sub.unsubscribe()
    }
  }, [sessionId, setGameState, setSelectedOption])

  // Determine which page to render
  function renderPage() {
    if (!sessionId) return <LobbyPage />

    switch (phase) {
      case undefined:
      case 'LOBBY':
        return <WaitingRoomPage />
      case 'CRISIS':
      case 'VOTING':
      case 'TALLY':
        return <VotingPage />
      case 'RESULT':
        return <ResultPage />
      case 'FINAL':
        return <FinalPage />
      default:
        return <WaitingRoomPage />
    }
  }

  return (
    <>
      {/* 3-D globe background */}
      <GlobeCanvas peaceScore={gameState?.peaceScore} />

      {/* HUD overlay */}
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '1rem',
        }}
      >
        {renderPage()}
      </div>
    </>
  )
}
