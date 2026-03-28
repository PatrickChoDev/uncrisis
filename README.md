# 🌐 UN Crisis Room

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
┌──────────────────────────────────────────────────────┐
│                    AWS Cloud                          │
│                                                      │
│  ┌──────────┐    ┌────────────┐   ┌───────────────┐  │
│  │CloudFront│───▶│     S3     │   │   AppSync     │  │
│  │  (CDN)   │    │ (Frontend) │   │ (GraphQL WS)  │  │
│  └──────────┘    └────────────┘   └───────┬───────┘  │
│                                           │           │
│  ┌───────────────────────────────────┐    │           │
│  │         ECS Fargate               │    │           │
│  │    Go Game Server                 │◀───┘           │
│  │  ┌──────────────┐ ┌────────────┐  │               │
│  │  │ Round Timer  │ │ SQS Poller │  │               │
│  │  └──────────────┘ └────┬───────┘  │               │
│  └───────────────────────────────────┘               │
│                           │                           │
│                    ┌──────▼──────┐                    │
│                    │     SQS     │                    │
│                    │ (Vote Queue)│                    │
│                    └─────────────┘                    │
│                                                      │
│  ┌──────────────────────────────────────────────────┐ │
│  │                  DynamoDB                        │ │
│  │  sessions-table  │  votes-table  │ scenarios-table│ │
│  └──────────────────────────────────────────────────┘ │
│                                                      │
│  ┌──────────┐                                        │
│  │   ECR    │  ← Docker image for the Go server     │
│  └──────────┘                                        │
└──────────────────────────────────────────────────────┘
```

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

### New Services

| # | Service | Provider | Role |
|---|---------|----------|------|
| 1 | **AWS ECS Fargate** | AWS | Persistent containerised game server — owns round timer, session state, SQS consumer |
| 2 | **Amazon SQS** | AWS | Vote queue — decouples vote ingestion from tally logic, handles burst traffic |
| 3 | **AWS AppSync** | AWS | Managed GraphQL API with WebSocket subscriptions — broadcasts real-time game state |
| 4 | **Amazon CloudFront** | AWS | CDN — serves React frontend from S3 with low latency |

### Supporting Services

| Service | Role |
|---------|------|
| Amazon ECR | Container registry for the Go game server Docker image |
| Amazon DynamoDB | Stores sessions, crisis scenarios, and vote records |
| Amazon S3 | Hosts compiled React frontend static assets |

---

## Repository Structure

```
uncrisis/
├── terraform/                  # AWS infrastructure (Terraform)
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
└── frontend/                   # React + Three.js SPA
    ├── src/
    │   ├── App.tsx             # Root component + AppSync subscription
    │   ├── config.ts           # Amplify / server config (env vars)
    │   ├── graphql.ts          # GraphQL queries / mutations / subscriptions
    │   ├── store.ts            # Zustand global state
    │   ├── types.ts            # TypeScript interfaces
    │   ├── styles.css          # Dark-theme global styles
    │   ├── components/
    │   │   ├── Globe.tsx       # Animated Three.js globe
    │   │   ├── GlobeCanvas.tsx # Full-screen R3F canvas
    │   │   ├── PlayerList.tsx  # Live player roster
    │   │   ├── Timer.tsx       # 30-second countdown bar
    │   │   └── ResultCard.tsx  # Round result + tally bars
    │   └── pages/
    │       ├── LobbyPage.tsx   # Create / join room
    │       ├── WaitingRoomPage.tsx
    │       ├── VotingPage.tsx  # Crisis card + vote buttons
    │       ├── ResultPage.tsx  # Round result reveal
    │       └── FinalPage.tsx   # Peace score ring
    ├── package.json
    └── vite.config.ts
```

---

## Prerequisites

| Tool | Minimum version |
|------|----------------|
| Node.js | 18 |
| Go | 1.22 |
| Terraform | 1.5 |
| AWS CLI | 2 (configured with deploy permissions) |
| Docker | 24 |

---

## Deployment

### 1 — Provision infrastructure

```bash
cd terraform
terraform init
terraform apply -var="project_name=uncrisis" -var="environment=prod"
```

Take note of the outputs:

```
cloudfront_url        = "https://dxxxxxxxxxx.cloudfront.net"
appsync_graphql_url   = "https://xxxxxxxxxx.appsync-api.ap-southeast-1.amazonaws.com/graphql"
appsync_api_key       = <sensitive>
ecr_repository_url    = "123456789.dkr.ecr.ap-southeast-1.amazonaws.com/uncrisis-game-server"
```

### 2 — Build & push the Go game server

```bash
cd backend

# Authenticate Docker to ECR
aws ecr get-login-password --region ap-southeast-1 \
  | docker login --username AWS --password-stdin <ECR_REPO_URL>

docker build -t uncrisis-game-server .
docker tag uncrisis-game-server:latest <ECR_REPO_URL>:latest
docker push <ECR_REPO_URL>:latest
```

ECS will automatically pull and restart the task.

### 3 — Build & deploy the React frontend

```bash
cd frontend

# Create a .env.local for local dev (never commit this file)
cat > .env.local <<EOF
VITE_APPSYNC_ENDPOINT=https://xxxxxxxxxx.appsync-api.ap-southeast-1.amazonaws.com/graphql
VITE_APPSYNC_API_KEY=<api_key_from_terraform_output>
VITE_APPSYNC_REGION=ap-southeast-1
VITE_GAME_SERVER_URL=http://<ecs-task-public-ip>:8080
EOF

npm install
npm run build

# Upload to S3 and invalidate CloudFront cache
BUCKET=$(terraform -chdir=../terraform output -raw s3_bucket_name)
DIST_ID=$(terraform -chdir=../terraform output -raw cloudfront_distribution_id)

aws s3 sync dist/ s3://$BUCKET/ --delete
aws cloudfront create-invalidation --distribution-id $DIST_ID --paths "/*"
```

---

## Local Development

### Backend

```bash
cd backend

export SQS_QUEUE_URL="http://localhost:4566/000000000000/uncrisis-votes"   # localstack
export APPSYNC_ENDPOINT="http://localhost:8081/graphql"
export APPSYNC_API_KEY="local-dev-key"
export AWS_REGION="ap-southeast-1"

go run ./cmd/server
```

### Frontend

```bash
cd frontend
cp .env.example .env.local   # fill in your values
npm run dev                  # http://localhost:5173
```

---

## Weekly Timeline

| Week | Milestone |
|------|-----------|
| 1 | Project setup, repository structure, Terraform skeleton |
| 2 | DynamoDB + SQS + ECR modules; Go project scaffolding |
| 3 | Go game manager (session lifecycle, round timer, vote tally) |
| 4 | AppSync schema, resolvers, real-time subscriptions |
| 5 | ECS Fargate task definition; Docker build pipeline |
| 6 | React + Three.js frontend: lobby, voting, result pages |
| 7 | CloudFront + S3 module; frontend build & deploy scripts |
| 8 | End-to-end integration testing; load testing SQS burst |
| 9 | UI polish, peace-score animations, scenario content |
| 10 | Documentation, final presentation, budget review |

---

## Estimated Budget (AWS, monthly)

| Service | Usage estimate | Cost (USD/mo) |
|---------|---------------|--------------|
| ECS Fargate | 1 task × 0.25 vCPU × 0.5 GB, 730 h | ~$9 |
| DynamoDB | On-demand, low traffic | ~$1 |
| SQS | <1M requests | <$1 |
| AppSync | <1M query/mutation/subscription units | ~$4 |
| CloudFront | 10 GB data transfer | ~$1 |
| S3 | 1 GB storage | <$1 |
| ECR | 1 GB storage | <$1 |
| **Total** | | **~$17 / month** |

---

## License

MIT © 2025 UN Crisis Room Team
