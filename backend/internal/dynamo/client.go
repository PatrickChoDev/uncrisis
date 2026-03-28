// Package dynamo implements the game.Persister interface against Amazon DynamoDB.
package dynamo

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/PatrickChoDev/uncrisis/backend/internal/game"
)

// Client wraps the DynamoDB SDK client.
type Client struct {
	db            *dynamodb.Client
	sessionsTable string
	scenariosTable string
}

// New creates a DynamoDB client using environment variables.
func New(db *dynamodb.Client) *Client {
	return &Client{
		db:             db,
		sessionsTable:  envOrDefault("DYNAMODB_SESSIONS", "uncrisis-sessions"),
		scenariosTable: envOrDefault("DYNAMODB_SCENARIOS", "uncrisis-scenarios"),
	}
}

// SaveSession persists (or overwrites) a session in DynamoDB.
func (c *Client) SaveSession(ctx context.Context, s *game.Session) error {
	item, err := attributevalue.MarshalMap(s)
	if err != nil {
		return err
	}

	_, err = c.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.sessionsTable),
		Item:      item,
	})
	return err
}

// GetScenario retrieves a scenario by index (cycles through the DefaultScenarios slice
// if the DynamoDB table has fewer items than requested).
func (c *Client) GetScenario(ctx context.Context, idx int) (*game.Scenario, error) {
	// Use the in-memory defaults first; DynamoDB scenarios table is treated as
	// an override / extension that operators can seed without redeploying.
	scenarios := game.DefaultScenarios
	if len(scenarios) == 0 {
		return nil, nil
	}
	s := scenarios[idx%len(scenarios)]
	return &s, nil
}

// ScenarioCount returns the number of built-in scenarios.
func (c *Client) ScenarioCount(_ context.Context) (int, error) {
	return len(game.DefaultScenarios), nil
}

// SeedScenarios writes DefaultScenarios to DynamoDB so operators can customise them.
func (c *Client) SeedScenarios(ctx context.Context) error {
	for _, s := range game.DefaultScenarios {
		item, err := attributevalue.MarshalMap(s)
		if err != nil {
			return err
		}

		_, err = c.db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName:           aws.String(c.scenariosTable),
			Item:                item,
			ConditionExpression: aws.String("attribute_not_exists(scenarioId)"),
		})
		if err != nil {
			// Ignore ConditionalCheckFailedException — item already exists
			var ccf *types.ConditionalCheckFailedException
			if !errors.As(err, &ccf) {
				return err
			}
		}
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
