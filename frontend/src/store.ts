import { create } from 'zustand'
import type { GameState, GamePhase } from './types'

interface GameStore {
  sessionId: string | null
  playerId: string | null
  displayName: string | null
  gameState: GameState | null
  selectedOption: string | null

  setSession: (sessionId: string, playerId: string, displayName: string) => void
  setGameState: (state: GameState) => void
  setSelectedOption: (optionId: string | null) => void
  reset: () => void
}

const initialState = {
  sessionId: null,
  playerId: null,
  displayName: null,
  gameState: null,
  selectedOption: null,
}

export const useGameStore = create<GameStore>((set) => ({
  ...initialState,

  setSession: (sessionId, playerId, displayName) =>
    set({ sessionId, playerId, displayName }),

  setGameState: (gameState) => set({ gameState }),

  setSelectedOption: (selectedOption) => set({ selectedOption }),

  reset: () => set(initialState),
}))

export type { GamePhase }
