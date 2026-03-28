FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /game-server ./cmd/server

# ── Runtime image ──────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12
COPY --from=builder /game-server /game-server
EXPOSE 8080
ENTRYPOINT ["/game-server"]
