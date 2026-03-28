// GraphQL type definitions mirroring the AppSync schema.

export type GamePhase =
  | 'LOBBY'
  | 'CRISIS'
  | 'VOTING'
  | 'TALLY'
  | 'RESULT'
  | 'FINAL'

export interface Option {
  id: string
  label: string
}

export interface Scenario {
  scenarioId: string
  title: string
  context: string
  options: Option[]
}

export interface Player {
  playerId: string
  displayName: string
  hasVoted: boolean
}

export interface OptionTally {
  optionId: string
  label: string
  count: number
  pct: number
}

export interface RoundResult {
  round: number
  tally: OptionTally[]
  winningOption: string
  consensusPct: number
  peaceScore: number
  narrativeOutcome: string
}

export interface GameState {
  sessionId: string
  phase: GamePhase
  currentRound: number
  totalRounds: number
  scenario?: Scenario
  players: Player[]
  timeLeftSecs: number
  lastResult?: RoundResult
  peaceScore: number
}
