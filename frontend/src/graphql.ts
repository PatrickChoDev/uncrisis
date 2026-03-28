import { generateClient } from 'aws-amplify/api'

const client = generateClient()

// ── Subscription ─────────────────────────────────────────────────────────────

const ON_GAME_STATE_UPDATED = /* GraphQL */ `
  subscription OnGameStateUpdated($sessionId: ID!) {
    onGameStateUpdated(sessionId: $sessionId) {
      sessionId
      phase
      currentRound
      totalRounds
      timeLeftSecs
      peaceScore
      scenario {
        scenarioId
        title
        context
        options { id label }
      }
      players {
        playerId
        displayName
        hasVoted
      }
      lastResult {
        round
        winningOption
        consensusPct
        peaceScore
        narrativeOutcome
        tally { optionId label count pct }
      }
    }
  }
`

// ── Mutation ──────────────────────────────────────────────────────────────────

const SUBMIT_VOTE = /* GraphQL */ `
  mutation SubmitVote($input: VoteInput!) {
    submitVote(input: $input) {
      sessionId
      playerId
      choice
      round
    }
  }
`

export function subscribeToGameState(
  sessionId: string,
  onData: (state: unknown) => void,
  onError: (err: unknown) => void,
) {
  const observable = client.graphql({
    query: ON_GAME_STATE_UPDATED,
    variables: { sessionId },
  })

  // client.graphql for subscriptions returns an Observable-like object
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const sub = (observable as any).subscribe({
    next: ({ data }: { data: { onGameStateUpdated: unknown } }) => {
      onData(data?.onGameStateUpdated)
    },
    error: onError,
  }) as { unsubscribe: () => void }

  return sub
}

export async function submitVote(
  sessionId: string,
  playerId: string,
  choice: string,
  round: number,
) {
  return client.graphql({
    query: SUBMIT_VOTE,
    variables: { input: { sessionId, playerId, choice, round } },
  })
}
