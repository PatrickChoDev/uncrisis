// Package sqsconsumer polls SQS for incoming votes and delivers them to the game manager.
package sqsconsumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/PatrickChoDev/uncrisis/backend/internal/game"
)

// VoteProcessor is called for every valid vote message.
type VoteProcessor interface {
	ProcessVote(ctx context.Context, vote game.Vote) error
}

// Consumer continuously long-polls an SQS queue and forwards votes.
type Consumer struct {
	client    *sqs.Client
	queueURL  string
	processor VoteProcessor
	logger    *slog.Logger
}

// New creates a Consumer.
func New(client *sqs.Client, queueURL string, processor VoteProcessor, logger *slog.Logger) *Consumer {
	return &Consumer{
		client:    client,
		queueURL:  queueURL,
		processor: processor,
		logger:    logger,
	}
}

// Run starts the polling loop; it blocks until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) {
	c.logger.Info("SQS consumer started", "queue", c.queueURL)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("SQS consumer stopping")
			return
		default:
		}

		msgs, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // long-poll
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("SQS receive error", "err", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, msg := range msgs.Messages {
			var vote game.Vote
			if err := json.Unmarshal([]byte(aws.ToString(msg.Body)), &vote); err != nil {
				c.logger.Warn("malformed vote message", "err", err, "body", aws.ToString(msg.Body))
				c.deleteMessage(ctx, msg.ReceiptHandle)
				continue
			}

			if err := c.processor.ProcessVote(ctx, vote); err != nil {
				c.logger.Error("process vote error", "err", err, "sessionId", vote.SessionID)
				// Leave message in-flight; it will become visible again after visibility timeout
				continue
			}

			c.deleteMessage(ctx, msg.ReceiptHandle)
		}
	}
}

func (c *Consumer) deleteMessage(ctx context.Context, receiptHandle *string) {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		c.logger.Warn("SQS delete error", "err", err)
	}
}
