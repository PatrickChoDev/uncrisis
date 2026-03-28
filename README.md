# UN Crisis Room

> Real-time multiplayer diplomacy game — powered by AWS, Go & React + Three.js

Players join a shared room and collectively vote on resolutions to global crisis scenarios. The game rewards **group consensus** and peaceful decision-making.

---

## Team

| Student ID | Name |
|------------|------|
| 6530075721 | Siriwid Thongon |
| 6530322721 | Phumsiri Sumativit |
| 6532089921 | Thanapat Chotipun |

---

## Problem Statement

Group debates and collective decision-making often end in circular arguments that consume time without producing actionable outcomes. Conventional communication tools lack structured mechanisms to guide participants toward resolution. There is a need for a structured, time-bounded platform that transforms debate into a game-like experience: one that is engaging, fair, and nudges groups toward **peaceful consensus** rather than prolonged conflict.

---

## Architecture

```
Browser
  │
  ├─── HTTPS /          ──▶ CloudFront ──▶ S3 (React SPA)
  │
  ├─── HTTPS /sessions* ──▶ CloudFront ──▶ ECS Fargate :8080 (Go game server)
  │                                          │  polls ──▶ SQS (vote queue)
  │                                          │  reads/writes ──▶ DynamoDB
  │                                          └─ mutations ──▶ AppSync
  │
  └─── WSS ─────────────▶ AppSync (GraphQL subscriptions)
                            └─ broadcasts game state to all clients
```

> CloudFront acts as a reverse proxy for the game server — the browser only ever speaks to `https://cloudfront-url`. No direct HTTP to ECS, no mixed-content errors.

### Game Flow

| Step | Phase | Description |
|------|-------|-------------|
| 1 | `LOBBY` | Players open the URL → enter room code + display name |
| 2 | `CRISIS` | Crisis scenario card appears: title, context, 4 options |
| 3 | `VOTING` | All players vote simultaneously (30-second window) |
| 4 | `TALLY` | ECS server tallies votes from SQS → computes consensus % |
| 5 | `RESULT` | AppSync broadcasts live result + narrative outcome |
| 6 | `FINAL` | After 5 rounds → collective Peace Score is shown |

---

## Cloud Services

| Service | Role |
|---------|------|
| **AWS ECS Fargate** | Persistent containerised game server — owns round timer, session state, SQS consumer |
| **Amazon SQS** | Vote queue — decouples vote ingestion from tally logic |
| **AWS AppSync** | Managed GraphQL API with WebSocket subscriptions — broadcasts real-time game state |
| **Amazon CloudFront** | CDN — serves React frontend from S3 with low latency |
| **Amazon ECR** | Container registry for the Go game server Docker image |
| **Amazon DynamoDB** | Stores sessions, crisis scenarios, and vote records |
| **Amazon S3** | Hosts compiled React frontend static assets |

---

## Repository Structure

```
uncrisis/
├── terraform/                  # AWS infrastructure (OpenTofu)
│   ├── main.tf                 # Root module: wires everything together
│   ├── variables.tf
│   ├── outputs.tf
│   └── modules/
│       ├── ecr/                # Container registry
│       ├── s3/                 # Frontend hosting bucket
│       ├── cloudfront/         # CDN distribution + OAC
│       ├── dynamodb/           # sessions / votes / scenarios tables
│       ├── sqs/                # vote queue + DLQ
│       ├── appsync/            # GraphQL API + schema + resolvers
│       └── ecs/                # Fargate cluster, task definition, service
│
├── backend/                    # Go game server
│   ├── cmd/server/main.go      # Entry point — HTTP API + wiring
│   ├── internal/
│   │   ├── game/               # Session manager, types, scenarios
│   │   ├── sqsconsumer/        # Long-poll SQS consumer loop
│   │   ├── appsync/            # AppSync HTTP mutation client
│   │   └── dynamo/             # DynamoDB persister
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/                   # React + Three.js SPA
│   ├── src/
│   │   ├── App.tsx             # Root component + AppSync subscription
│   │   ├── config.ts           # Amplify / server config (env vars)
│   │   ├── graphql.ts          # GraphQL queries / mutations / subscriptions
│   │   ├── store.ts            # Zustand global state
│   │   ├── components/
│   │   │   ├── Globe.tsx       # Animated Three.js globe
│   │   │   ├── PlayerList.tsx  # Live player roster
│   │   │   ├── Timer.tsx       # 30-second countdown bar
│   │   │   └── ResultCard.tsx  # Round result + tally bars
│   │   └── pages/
│   │       ├── LobbyPage.tsx
│   │       ├── WaitingRoomPage.tsx
│   │       ├── VotingPage.tsx
│   │       ├── ResultPage.tsx
│   │       └── FinalPage.tsx
│   ├── package.json
│   └── vite.config.ts
│
└── scripts/
    ├── deploy-backend.sh       # Build → push → ECS redeploy → update CloudFront origin
    ├── deploy-frontend.sh      # npm build → S3 sync → CloudFront invalidation
    ├── start-server.sh         # Scale ECS to 1 → wait → update CloudFront origin
    └── stop-server.sh          # Scale ECS to 0 (stops Fargate charges)
```

---

## Prerequisites

| Tool | Minimum version |
|------|----------------|
| Node.js | 18 |
| Go | 1.22 |
| OpenTofu | 1.6 |
| AWS CLI | 2 (configured with deploy permissions) |
| Docker | 24 |

---

## Deployment

### First-time setup (run once)

#### 1 — Provision infrastructure

```bash
cd terraform
tofu init
tofu apply -var="environment=prod"
```

Note the sensitive API key output:

```bash
tofu output appsync_api_key
```

All outputs available after apply:

```
cloudfront_url              = "https://dxxxxxxxxxx.cloudfront.net"
cloudfront_distribution_id  = "EXXXXXXXXXXXX"
appsync_graphql_url         = "https://xxxxxxxxxx.appsync-api.ap-southeast-1.amazonaws.com/graphql"
appsync_api_key             = <sensitive>
ecr_repository_url          = "xxxxxxxxxxx.dkr.ecr.ap-southeast-1.amazonaws.com/uncrisis-game-server"
ecs_cluster_name            = "uncrisis-cluster"
ecs_service_name            = "uncrisis-game-server"
s3_bucket_name              = "uncrisis-frontend-prod"
sqs_vote_queue_url          = "https://sqs.ap-southeast-1.amazonaws.com/..."
```

#### 2 — Deploy the backend

```bash
./scripts/deploy-backend.sh
```

This single command:
1. Builds the Docker image and pushes to ECR
2. Forces an ECS service redeployment
3. Waits for the new task to be running (~60-90s)
4. Discovers the new task's public IP via AWS CLI
5. Updates the **CloudFront ECS origin** to the new IP (propagates in ~3-5 min)

The game server is always accessed through CloudFront (`https://cloudfront_url/sessions`), never directly by IP. No `.env.local` changes are needed between deploys.

#### 3 — Deploy the frontend (first time only)

```bash
./scripts/deploy-frontend.sh
```

This script creates `frontend/.env.local` automatically from Terraform outputs (AppSync endpoint, API key, and CloudFront URL as `VITE_GAME_SERVER_URL`), then builds and deploys to S3 + CloudFront.

---

### Subsequent deploys

| Change | Command |
|--------|---------|
| Backend code changed | `./scripts/deploy-backend.sh` — updates ECS + CloudFront origin |
| Frontend only changed | `./scripts/deploy-frontend.sh` |
| Before a game session | `./scripts/start-server.sh` — scales ECS to 1, updates CloudFront origin |
| After a game session | `./scripts/stop-server.sh` — scales ECS to 0, stops Fargate charges |

---

## Local Development

### Backend

```bash
cd backend

export SQS_QUEUE_URL="http://localhost:4566/000000000000/uncrisis-votes"
export APPSYNC_ENDPOINT="http://localhost:8081/graphql"
export APPSYNC_API_KEY="local-dev-key"
export AWS_REGION="ap-southeast-1"

go run ./cmd/server
```

### Frontend

```bash
cd frontend

# Create local env pointing at localhost backend
cat > .env.local <<EOF
VITE_APPSYNC_ENDPOINT=http://localhost:8081/graphql
VITE_APPSYNC_API_KEY=local-dev-key
VITE_APPSYNC_REGION=ap-southeast-1
VITE_GAME_SERVER_URL=http://localhost:8080
EOF

npm install
npm run dev   # http://localhost:5173
```

---

## Estimated Budget (AWS, monthly)

Assuming **4 active hours/day** (run `start-server.sh` before sessions, `stop-server.sh` after):

| Service | Usage estimate | Cost (USD/mo) |
|---------|---------------|--------------|
| ECS Fargate Spot | 0.25 vCPU × 0.5 GB × ~120 h/mo | ~$0.50 |
| AppSync | <1M query/mutation/subscription units | ~$2 |
| DynamoDB | On-demand, 50 users | ~$0.50 |
| CloudFront | 10 GB data transfer | ~$1 |
| SQS | <1M requests | <$1 |
| S3 | 1 GB storage | <$1 |
| ECR | <500 MB (3 images) | free tier |
| CloudWatch | 3-day log retention | <$0.10 |
| **Total** | | **~$4 / month** |

> If the server runs 24/7 (desired_count always 1), ECS cost rises to ~$3/month. Still well under $10/month total.

---

## License

MIT © 2025 UN Crisis Room Team
