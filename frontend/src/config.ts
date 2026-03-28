// Amplify / AppSync configuration.
// Values are injected at build time via environment variables:
//   VITE_APPSYNC_ENDPOINT   — GraphQL HTTPS URL
//   VITE_APPSYNC_API_KEY    — AppSync API key
//   VITE_APPSYNC_REGION     — AWS region (default: ap-southeast-1)

export const amplifyConfig = {
  API: {
    GraphQL: {
      endpoint: import.meta.env.VITE_APPSYNC_ENDPOINT as string,
      region: (import.meta.env.VITE_APPSYNC_REGION as string) || 'ap-southeast-1',
      defaultAuthMode: 'apiKey' as const,
      apiKey: import.meta.env.VITE_APPSYNC_API_KEY as string,
    },
  },
}

export const GAME_SERVER_URL =
  (import.meta.env.VITE_GAME_SERVER_URL as string) || 'http://localhost:8080'
