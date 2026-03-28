// Command server starts the UN Crisis Room game server.
// It wires together the game manager, SQS consumer, AppSync notifier,
// and an HTTP API for room management.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	appsyncpkg "github.com/PatrickChoDev/uncrisis/backend/internal/appsync"
	"github.com/PatrickChoDev/uncrisis/backend/internal/dynamo"
	"github.com/PatrickChoDev/uncrisis/backend/internal/game"
	"github.com/PatrickChoDev/uncrisis/backend/internal/sqsconsumer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// ── Configuration ────────────────────────────────────────────────────────
	roundDuration := envDuration("ROUND_DURATION_SECS", 30) * time.Second
	totalRounds := envInt("TOTAL_ROUNDS", 5)
	sqsQueueURL := mustEnv("SQS_QUEUE_URL")
	appsyncEndpoint := mustEnv("APPSYNC_ENDPOINT")
	appsyncAPIKey := mustEnv("APPSYNC_API_KEY")
	httpAddr := envOrDefault("HTTP_ADDR", ":8080")

	// ── AWS SDK ───────────────────────────────────────────────────────────────
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load AWS config", "err", err)
		os.Exit(1)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	sqsClient := awssqs.NewFromConfig(cfg)

	// ── Clients & manager ────────────────────────────────────────────────────
	dbClient := dynamo.New(dynamoClient)
	notifier := appsyncpkg.New(appsyncEndpoint, appsyncAPIKey)
	mgr := game.NewManager(dbClient, notifier, roundDuration, totalRounds, logger)

	// Seed built-in scenarios (idempotent)
	if err := dbClient.SeedScenarios(ctx); err != nil {
		logger.Warn("scenario seed warning (may already exist)", "err", err)
	}

	// ── SQS consumer ─────────────────────────────────────────────────────────
	consumer := sqsconsumer.New(sqsClient, sqsQueueURL, mgr, logger)
	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	go consumer.Run(consumerCtx)

	// ── HTTP API ──────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// POST /sessions — create a new game room
	mux.HandleFunc("POST /sessions", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			HostName    string `json:"hostName"`
			TotalRounds int    `json:"totalRounds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.HostName == "" {
			http.Error(w, "hostName required", http.StatusBadRequest)
			return
		}
		if req.TotalRounds <= 0 {
			req.TotalRounds = totalRounds
		}

		sessionID, err := mgr.CreateSession(r.Context(), req.HostName)
		if err != nil {
			logger.Error("create session", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]string{"sessionId": sessionID})
	})

	// POST /sessions/{sessionId}/join — player joins a lobby
	mux.HandleFunc("POST /sessions/{sessionId}/join", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("sessionId")
		var req struct {
			PlayerID    string `json:"playerId"`
			DisplayName string `json:"displayName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := mgr.JoinSession(r.Context(), sessionID, req.PlayerID, req.DisplayName); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// POST /sessions/{sessionId}/start — host starts the next round
	mux.HandleFunc("POST /sessions/{sessionId}/start", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("sessionId")
		if err := mgr.StartRound(r.Context(), sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	go func() {
		logger.Info("HTTP server listening", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	cancelConsumer()

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable not set", "key", key)
		os.Exit(1)
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func envDuration(key string, defSecs int) time.Duration {
	return time.Duration(envInt(key, defSecs))
}
